/*
Copyright 2021 Red Hat OpenShift Data Foundation.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	"github.com/red-hat-storage/odf-operator/pkg/util"
)

// CheckForExistingSubscription looks for any existing Subscriptions that
// reference the given package. If one does exist, use its ObjectMeta for the
// desiredSubscription.
//
// NOTE(jarrpa): We can't use client.MatchingFields to limit the list results
// because fake.Client does not support them.
func CheckExistingSubscriptions(cli client.Client, desiredSubscription *operatorv1alpha1.Subscription) (*operatorv1alpha1.Subscription, error) {

	subsList := &operatorv1alpha1.SubscriptionList{}
	err := cli.List(context.TODO(), subsList, &client.ListOptions{Namespace: desiredSubscription.Namespace})
	if err != nil {
		return nil, err
	}

	var actualSub *operatorv1alpha1.Subscription
	pkgName := desiredSubscription.Spec.Package
	for i, sub := range subsList.Items {
		if sub.Spec.Package == pkgName {
			if actualSub != nil {
				foundSubs := []string{actualSub.Name, sub.Name}
				return nil, fmt.Errorf("multiple Subscriptions found for package '%s': %v", pkgName, foundSubs)
			}
			actualSub = &subsList.Items[i]
			actualSub.Spec.Channel = desiredSubscription.Spec.Channel

			// If the config is not set, only then set it to the desired value => allow user to override
			if actualSub.Spec.Config == nil {
				actualSub.Spec.Config = desiredSubscription.Spec.Config
			} else if actualSub.Spec.Config != nil && desiredSubscription.Spec.Config != nil {
				// Combines the environment variables from both subscriptions.
				// If actualSub already contains an environment variable, its value will be updated with the value from desiredSubscription.
				actualSub.Spec.Config.Env = getMergedEnvVars(actualSub.Spec.Config.Env, desiredSubscription.Spec.Config.Env)
			}

			desiredSubscription = actualSub
		}
	}

	return desiredSubscription, nil
}

// getMergedEnvVars updates the value of env variables in the envList1 with that of envList2 and
// returns the updated list of env variables.
func getMergedEnvVars(envList1, envList2 []corev1.EnvVar) []corev1.EnvVar {
	envMap := make(map[string]string)

	for _, env := range envList1 {
		envMap[env.Name] = env.Value
	}

	for _, env := range envList2 {
		envMap[env.Name] = env.Value
	}

	// Convert the map back to a slice
	var updatedEnvVars []corev1.EnvVar
	for key, value := range envMap {
		updatedEnvVars = append(updatedEnvVars, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return updatedEnvVars
}

func EnsureDesiredSubscription(cli client.Client, desiredSubscription *operatorv1alpha1.Subscription) error {

	var err error

	desiredSubscription, err = CheckExistingSubscriptions(cli, desiredSubscription)
	if err != nil {
		return err
	}

	// create/update subscription
	sub := &operatorv1alpha1.Subscription{}
	sub.ObjectMeta = desiredSubscription.ObjectMeta
	_, err = controllerutil.CreateOrUpdate(context.TODO(), cli, sub, func() error {
		sub.Spec = desiredSubscription.Spec
		return SetOdfSubControllerReference(cli, sub)
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func SetOdfSubControllerReference(cli client.Client, obj client.Object) error {

	odfSub, err := GetOdfSubscription(cli)
	if err != nil {
		return err
	}

	err = controllerutil.SetControllerReference(odfSub, obj, cli.Scheme())
	if err != nil {
		return err
	}

	return nil
}

// GetOdfSubscription returns the subscription for odf-operator.
func GetOdfSubscription(cli client.Client) (*operatorv1alpha1.Subscription, error) {

	subsList := &operatorv1alpha1.SubscriptionList{}
	err := cli.List(context.TODO(), subsList, &client.ListOptions{Namespace: OperatorNamespace})
	if err != nil {
		return nil, err
	}

	for _, sub := range subsList.Items {
		if sub.Spec.Package == OdfSubscriptionPackage {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("odf-operator subscription not found")
}

func GetVendorCsvNames(kind odfv1alpha1.StorageKind) []string {

	var csvNames []string

	if kind == VendorFlashSystemCluster() {
		csvNames = []string{IbmSubscriptionStartingCSV}
	} else if kind == VendorStorageCluster() {
		csvNames = []string{NoobaaSubscriptionStartingCSV, OcsSubscriptionStartingCSV, CSIAddonsSubscriptionStartingCSV}
	}

	return csvNames
}

func EnsureVendorCsv(cli client.Client, csvName string) (*operatorv1alpha1.ClusterServiceVersion, error) {

	var err error

	csvObj := &operatorv1alpha1.ClusterServiceVersion{}
	err = cli.Get(context.TODO(), types.NamespacedName{
		Name: csvName, Namespace: OperatorNamespace}, csvObj)
	if err != nil {
		if errors.IsNotFound(err) {
			approvalErr := ApproveInstallPlanForCsv(cli, csvName)
			if approvalErr != nil {
				return nil, approvalErr
			}
		}
		return nil, err
	}
	_, err = controllerutil.CreateOrUpdate(context.TODO(), cli, csvObj, func() error {
		csvObj.OwnerReferences = []metav1.OwnerReference{}
		return SetOdfSubControllerReference(cli, csvObj)
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}

	isReady := csvObj.Status.Phase == operatorv1alpha1.CSVPhaseSucceeded &&
		csvObj.Status.Reason == operatorv1alpha1.CSVReasonInstallSuccessful

	if !isReady {
		err = fmt.Errorf("CSV is not successfully installed")
		return nil, err
	}

	return csvObj, err
}

// ApproveInstallPlanForCsv approve the manual approval installPlan for the given CSV
// and returns an error if none found
func ApproveInstallPlanForCsv(cli client.Client, csvName string) error {

	var finalError error
	var foundInstallPlan bool

	installPlans := &operatorv1alpha1.InstallPlanList{}
	err := cli.List(context.TODO(), installPlans, &client.ListOptions{Namespace: OperatorNamespace})

	if err != nil {
		return err
	}

	for i, installPlan := range installPlans.Items {
		if util.FindInSlice(installPlan.Spec.ClusterServiceVersionNames, csvName) {
			foundInstallPlan = true
			if installPlan.Spec.Approval == operatorv1alpha1.ApprovalManual &&
				!installPlan.Spec.Approved {

				installPlans.Items[i].Spec.Approved = true
				err = cli.Update(context.TODO(), &installPlans.Items[i])
				if err != nil {
					multierr.AppendInto(&finalError, fmt.Errorf(
						"Failed to approve installplan %s", installPlan.Name))
					multierr.AppendInto(&finalError, err)
				}
			}
		}
	}

	if !foundInstallPlan {
		err = fmt.Errorf("InstallPlan not found for CSV %s", csvName)
		multierr.AppendInto(&finalError, err)
	}

	return finalError
}

// GetSubscriptions returns all required Subscriptions for the given StorageKind
func GetSubscriptions(k odfv1alpha1.StorageKind) []*operatorv1alpha1.Subscription {

	subscriptions := []*operatorv1alpha1.Subscription{}
	if k == StorageClusterKind {
		subscriptions = GetStorageClusterSubscriptions()
	} else if k == FlashSystemKind {
		subscriptions = GetFlashSystemClusterSubscriptions()
	}

	return subscriptions
}

// GetStorageClusterSubscription return subscription for StorageCluster
func GetStorageClusterSubscriptions() []*operatorv1alpha1.Subscription {
	noobaaSubscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NoobaaSubscriptionName,
			Namespace: OperatorNamespace,
		},
		Spec: &operatorv1alpha1.SubscriptionSpec{
			CatalogSource:          NoobaaSubscriptionCatalogSource,
			CatalogSourceNamespace: NoobaaSubscriptionCatalogSourceNamespace,
			Package:                NoobaaSubscriptionPackage,
			Channel:                NoobaaSubscriptionChannel,
			StartingCSV:            NoobaaSubscriptionStartingCSV,
			InstallPlanApproval:    operatorv1alpha1.ApprovalAutomatic,
		},
	}

	roleARN := os.Getenv("ROLEARN")
	if roleARN != "" {
		noobaaSubscription.Spec.Config = &operatorv1alpha1.SubscriptionConfig{
			Env: []corev1.EnvVar{
				{
					Name:  "ROLEARN",
					Value: roleARN,
				},
			},
		}
	}

	ocsSubscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      OcsSubscriptionName,
			Namespace: OperatorNamespace,
		},
		Spec: &operatorv1alpha1.SubscriptionSpec{
			CatalogSource:          OcsSubscriptionCatalogSource,
			CatalogSourceNamespace: OcsSubscriptionCatalogSourceNamespace,
			Package:                OcsSubscriptionPackage,
			Channel:                OcsSubscriptionChannel,
			StartingCSV:            OcsSubscriptionStartingCSV,
			InstallPlanApproval:    operatorv1alpha1.ApprovalAutomatic,
		},
	}

	csiAddonsSubscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CSIAddonsSubscriptionName,
			Namespace: OperatorNamespace,
		},
		Spec: &operatorv1alpha1.SubscriptionSpec{
			CatalogSource:          CSIAddonsSubscriptionCatalogSource,
			CatalogSourceNamespace: CSIAddonsSubscriptionCatalogSourceNamespace,
			Package:                CSIAddonsSubscriptionPackage,
			Channel:                CSIAddonsSubscriptionChannel,
			StartingCSV:            CSIAddonsSubscriptionStartingCSV,
			InstallPlanApproval:    operatorv1alpha1.ApprovalAutomatic,
			Config: &operatorv1alpha1.SubscriptionConfig{
				Tolerations: []corev1.Toleration{
					{
						Key:      "node.ocs.openshift.io/storage",
						Operator: "Equal",
						Value:    "true",
						Effect:   "NoSchedule",
					},
				},
			},
		},
	}

	return []*operatorv1alpha1.Subscription{noobaaSubscription, ocsSubscription, csiAddonsSubscription}
}

// GetFlashSystemClusterSubscription return subscription for FlashSystemCluster
func GetFlashSystemClusterSubscriptions() []*operatorv1alpha1.Subscription {
	ibmSubscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      IbmSubscriptionName,
			Namespace: OperatorNamespace,
		},
		Spec: &operatorv1alpha1.SubscriptionSpec{
			CatalogSource:          IbmSubscriptionCatalogSource,
			CatalogSourceNamespace: IbmSubscriptionCatalogSourceNamespace,
			Package:                IbmSubscriptionPackage,
			Channel:                IbmSubscriptionChannel,
			StartingCSV:            IbmSubscriptionStartingCSV,
			InstallPlanApproval:    operatorv1alpha1.ApprovalAutomatic,
		},
	}

	return []*operatorv1alpha1.Subscription{ibmSubscription}
}
