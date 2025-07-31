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
	"slices"

	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// CheckForExistingSubscription looks for any existing Subscriptions that
// reference the given package. If one does exist, use its ObjectMeta for the
// desiredSubscription.
//
// NOTE(jarrpa): We can't use client.MatchingFields to limit the list results
// because fake.Client does not support them.
func GetDesiredSubscription(ctx context.Context, cli client.Client, record *OlmPkgRecord) (*opv1a1.Subscription, error) {

	desiredSubscription := &opv1a1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      record.Pkg,
			Namespace: record.Namespace,
		},
		Spec: &opv1a1.SubscriptionSpec{
			Package:     record.Pkg,
			Channel:     record.Channel,
			StartingCSV: record.Csv,
		},
	}

	AdjustSpecialCasesSubscriptionConfig(desiredSubscription)

	odfSub, err := GetOdfSubscription(ctx, cli)
	if err != nil {
		return nil, err
	}

	if odfSub.Spec.Config == nil {
		odfSub.Spec.Config = &opv1a1.SubscriptionConfig{}
	}

	subsList := &opv1a1.SubscriptionList{}
	err = cli.List(ctx, subsList, &client.ListOptions{Namespace: desiredSubscription.Namespace})
	if err != nil {
		return nil, err
	}

	var subExsist bool
	var actualSub *opv1a1.Subscription
	pkgName := desiredSubscription.Spec.Package
	for i, sub := range subsList.Items {
		if sub.Spec.Package == pkgName {
			subExsist = true
			if actualSub != nil {
				foundSubs := []string{actualSub.Name, sub.Name}
				return nil, fmt.Errorf("multiple Subscriptions found for package '%s': %v", pkgName, foundSubs)
			}
			actualSub = &subsList.Items[i]

			actualSub.Spec.Channel = desiredSubscription.Spec.Channel
			if actualSub.Spec.Config == nil && desiredSubscription.Spec.Config == nil {
				actualSub.Spec.Config = &opv1a1.SubscriptionConfig{}
				actualSub.Spec.Config.Tolerations = odfSub.Spec.Config.Tolerations
			} else if actualSub.Spec.Config == nil && desiredSubscription.Spec.Config != nil {
				actualSub.Spec.Config = desiredSubscription.Spec.Config
				actualSub.Spec.Config.Tolerations = getMergedTolerations(odfSub.Spec.Config.Tolerations, desiredSubscription.Spec.Config.Tolerations)
			} else if actualSub.Spec.Config != nil && desiredSubscription.Spec.Config == nil {
				actualSub.Spec.Config.Tolerations = odfSub.Spec.Config.Tolerations
			} else if actualSub.Spec.Config != nil && desiredSubscription.Spec.Config != nil {
				// Combines the environment variables from both subscriptions.
				actualSub.Spec.Config.Env = getMergedEnvVars(actualSub.Spec.Config.Env, desiredSubscription.Spec.Config.Env)
				// Combines the Tolerations from odf sub and desired sub.
				actualSub.Spec.Config.Tolerations = getMergedTolerations(odfSub.Spec.Config.Tolerations, desiredSubscription.Spec.Config.Tolerations)
			}

			desiredSubscription = actualSub
		}
	}

	if !subExsist {
		// Set the catalog source for the odf-dependencies subscription to match that of the odf-operator subscription
		// This ensures that the odf-dependencies subscription uses the same catalog source across all environments,
		// including offline and test environments where the catalog name may vary.
		if desiredSubscription.Spec.Package == OdfDepsSubscriptionPackage {
			desiredSubscription.Spec.CatalogSource = odfSub.Spec.CatalogSource
			desiredSubscription.Spec.CatalogSourceNamespace = odfSub.Spec.CatalogSourceNamespace
		}

		if desiredSubscription.Spec.Config == nil {
			desiredSubscription.Spec.Config = &opv1a1.SubscriptionConfig{
				Tolerations: odfSub.Spec.Config.Tolerations,
			}
		} else {
			desiredSubscription.Spec.Config.Tolerations = getMergedTolerations(odfSub.Spec.Config.Tolerations, desiredSubscription.Spec.Config.Tolerations)
		}
	}

	return desiredSubscription, nil
}

func getMergedTolerations(tol1, tol2 []corev1.Toleration) []corev1.Toleration {

	if len(tol1) == 0 {
		return append([]corev1.Toleration{}, tol2...)
	} else if len(tol2) == 0 {
		return append([]corev1.Toleration{}, tol1...)
	}

	mergedTolerations := append([]corev1.Toleration{}, tol1...)

	for _, t2 := range tol2 {
		found := false
		for i, t1 := range tol1 {
			if t1.Key == t2.Key && t1.Operator == t2.Operator && t1.Value == t2.Value {
				found = true
				break
			}
			// If the toleration with the same key but different values is found,
			// update the existing toleration in tol1 with the new toleration from tol2.
			if t1.Key == t2.Key && t1.Operator == t2.Operator {
				tol1[i] = t2
				found = true
				break
			}
		}
		if !found {
			mergedTolerations = append(mergedTolerations, t2)
		}
	}

	return mergedTolerations
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

func EnsureDesiredSubscription(ctx context.Context, cli client.Client, olmPkgRecord *OlmPkgRecord) error {

	var err error

	desiredSubscription, err := GetDesiredSubscription(ctx, cli, olmPkgRecord)
	if err != nil {
		return err
	}

	// Skip creating (only update) subscriptions other than odf-dependencies
	// It will allow OLM to manage their creation via dependency resolution
	if desiredSubscription.Spec.Package != OdfDepsSubscriptionPackage && desiredSubscription.CreationTimestamp.IsZero() {
		return nil
	}

	// create/update subscription
	sub := &opv1a1.Subscription{}
	sub.ObjectMeta = desiredSubscription.ObjectMeta
	_, err = controllerutil.CreateOrUpdate(ctx, cli, sub, func() error {
		sub.Spec = desiredSubscription.Spec

		if desiredSubscription.Namespace == OperatorNamespace {
			return SetOdfSubControllerReference(ctx, cli, sub)
		}

		return nil
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func SetOdfSubControllerReference(ctx context.Context, cli client.Client, obj client.Object) error {

	odfSub, err := GetOdfSubscription(ctx, cli)
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
func GetOdfSubscription(ctx context.Context, cli client.Client) (*opv1a1.Subscription, error) {

	subsList := &opv1a1.SubscriptionList{}
	err := cli.List(ctx, subsList, &client.ListOptions{Namespace: OperatorNamespace})
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

func EnsureCsv(ctx context.Context, cli client.Client, olmPkgRecord *OlmPkgRecord) error {

	csvObj := &opv1a1.ClusterServiceVersion{}
	csvObj.Name, csvObj.Namespace = olmPkgRecord.Csv, olmPkgRecord.Namespace

	if err := cli.Get(ctx, client.ObjectKeyFromObject(csvObj), csvObj); err != nil {
		if errors.IsNotFound(err) {
			if present, err := isSubscriptionPresent(ctx, cli, olmPkgRecord); err != nil {
				return err
			} else if present {
				if err := ApproveInstallPlanForCsv(ctx, cli, olmPkgRecord.Csv, olmPkgRecord.Namespace); err != nil {
					return err
				}
			}
			return nil
		}
		return err
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, cli, csvObj, func() error {

		if csvObj.Namespace == OperatorNamespace {
			return SetOdfSubControllerReference(ctx, cli, csvObj)
		}

		return nil
	}); err != nil {
		return err
	}

	isReady := csvObj.Status.Phase == opv1a1.CSVPhaseSucceeded &&
		csvObj.Status.Reason == opv1a1.CSVReasonInstallSuccessful

	if !isReady {
		return fmt.Errorf("CSV is not successfully installed")
	}

	return nil
}

func isSubscriptionPresent(ctx context.Context, cli client.Client, olmPkgRecord *OlmPkgRecord) (bool, error) {

	// get all subscriptions in the cluster
	subList := &opv1a1.SubscriptionList{}
	if err := cli.List(ctx, subList, client.InNamespace(olmPkgRecord.Namespace)); err != nil {
		return false, err
	}

	// check if subscription exists on cluster for the given csv
	for _, sub := range subList.Items {
		if sub.Spec.Package == olmPkgRecord.Pkg {
			return true, nil
		}
	}

	return false, nil
}

// ApproveInstallPlanForCsv approve the manual approval installPlan for the given CSV
// and returns an error if none found
func ApproveInstallPlanForCsv(ctx context.Context, cli client.Client, csvName string, namespace string) error {

	var finalError error
	var foundInstallPlan bool

	installPlans := &opv1a1.InstallPlanList{}
	err := cli.List(ctx, installPlans, &client.ListOptions{Namespace: namespace})

	if err != nil {
		return err
	}

	for i, installPlan := range installPlans.Items {
		if slices.Contains(installPlan.Spec.ClusterServiceVersionNames, csvName) {
			foundInstallPlan = true
			if installPlan.Spec.Approval == opv1a1.ApprovalManual &&
				!installPlan.Spec.Approved {

				installPlans.Items[i].Spec.Approved = true
				err = cli.Update(ctx, &installPlans.Items[i])
				if err != nil {
					multierr.AppendInto(&finalError, fmt.Errorf(
						"failed to approve installplan %s", installPlan.Name))
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

func AdjustSpecialCasesSubscriptionConfig(subscription *opv1a1.Subscription) {

	switch subscription.Spec.Package {

	case "csi-addons", "odf-csi-addons-operator", "cephcsi-operator":
		subscription.Spec.Config = &opv1a1.SubscriptionConfig{
			Tolerations: []corev1.Toleration{
				{
					Key:      "node.ocs.openshift.io/storage",
					Operator: "Equal",
					Value:    "true",
					Effect:   "NoSchedule",
				},
			},
		}

	case "noobaa-operator", "mcg-operator":
		roleARN := os.Getenv("ROLEARN")
		if roleARN != "" {
			subscription.Spec.Config = &opv1a1.SubscriptionConfig{
				Env: []corev1.EnvVar{
					{
						Name:  "ROLEARN",
						Value: roleARN,
					},
				},
			}
		}
	}
}
