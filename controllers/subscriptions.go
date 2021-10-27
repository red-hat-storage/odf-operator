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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
)

// CheckForExistingSubscription looks for any existing Subscriptions that
// reference the given package. If one does exist, use its ObjectMeta for the
// desiredSubscription.
//
// NOTE(jarrpa): We can't use client.MatchingFields to limit the list results
// because fake.Client does not support them.
func CheckExistingSubscriptions(cli client.Client, desiredSubscription *operatorv1alpha1.Subscription) (*operatorv1alpha1.Subscription, error) {

	subsList := &operatorv1alpha1.SubscriptionList{}
	err := cli.List(context.TODO(), subsList)
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
			desiredSubscription = actualSub
		}
	}

	return desiredSubscription, nil
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

func GetVendorCsvNames(kind odfv1alpha1.StorageKind) []string {

	var csvNames []string

	if kind == VendorFlashSystemCluster() {
		csvNames = []string{IbmSubscriptionStartingCSV}
	} else if kind == VendorStorageCluster() {
		csvNames = []string{NoobaaSubscriptionStartingCSV, OcsSubscriptionStartingCSV}
	}

	return csvNames
}

func EnsureVendorCsv(cli client.Client, csvName string) (*operatorv1alpha1.ClusterServiceVersion, error) {

	var err error

	csvObj := &operatorv1alpha1.ClusterServiceVersion{}
	err = cli.Get(context.TODO(), types.NamespacedName{
		Name: csvName, Namespace: OperatorNamespace}, csvObj)
	if err != nil {
		return nil, err
	}

	_, err = controllerutil.CreateOrUpdate(context.TODO(), cli, csvObj, func() error {
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

// GetOdfSubscription return subscription for odf-operator and store it in cache
// It fetch once and use the same again and again from cache
func GetOdfSubscription(cli client.Client) (*operatorv1alpha1.Subscription, error) {

	if OdfSubscriptionObjectMeta != nil {
		return &operatorv1alpha1.Subscription{ObjectMeta: *OdfSubscriptionObjectMeta}, nil
	}

	odfSub := &operatorv1alpha1.Subscription{}

	err := cli.Get(context.TODO(), types.NamespacedName{
		Name: OdfSubscriptionName, Namespace: OperatorNamespace}, odfSub)
	if err != nil {
		return nil, err
	}

	OdfSubscriptionObjectMeta = &odfSub.ObjectMeta

	return &operatorv1alpha1.Subscription{ObjectMeta: *OdfSubscriptionObjectMeta}, nil
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

	return []*operatorv1alpha1.Subscription{ocsSubscription, noobaaSubscription}
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
