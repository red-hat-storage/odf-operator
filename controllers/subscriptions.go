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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
)

// CheckForExistingSubscription looks for any existing Subscriptions that
// reference the given package.
//
// NOTE(jarrpa): We can't use client.MatchingFields to limit the list results
// because fake.Client does not support them.
func (r *StorageSystemReconciler) CheckExistingSubscriptions(desiredSubscription *operatorv1alpha1.Subscription) error {

	subsList := &operatorv1alpha1.SubscriptionList{}
	err := r.Client.List(context.TODO(), subsList)
	if err != nil {
		return err
	}

	var actualSub *operatorv1alpha1.Subscription
	pkgName := desiredSubscription.Spec.Package
	for i, sub := range subsList.Items {
		if sub.Spec.Package == pkgName {
			if actualSub != nil {
				foundSubs := []string{actualSub.Name, sub.Name}
				return fmt.Errorf("multiple Subscriptions found for package '%s': %v", pkgName, foundSubs)
			}
			if sub.Name != desiredSubscription.Name {
				return fmt.Errorf("found conflicting Subscription found for package '%s': %s", pkgName, sub.Name)
			}
			actualSub = &subsList.Items[i]
		}
	}

	return nil
}

func (r *StorageSystemReconciler) ensureSubscription(instance *odfv1alpha1.StorageSystem, logger logr.Logger) error {

	var err error

	subs := GetSubscriptions(instance.Spec.Kind)
	if len(subs) == 0 {
		return fmt.Errorf("no subscriptions defined for kind: %v", instance.Spec.Kind)
	}

	for _, desiredSubscription := range subs {
		err = r.CheckExistingSubscriptions(desiredSubscription)
		if err != nil {
			return err
		}

		// create/update subscription
		sub := &operatorv1alpha1.Subscription{}
		sub.ObjectMeta = desiredSubscription.ObjectMeta
		_, err = controllerutil.CreateOrUpdate(context.TODO(), r.Client, sub, func() error {
			sub.Spec = desiredSubscription.Spec
			return controllerutil.SetControllerReference(instance, sub, r.Scheme)
		})
		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Error(err, "failed to create subscription")
			return err
		}

		err = r.addReferenceToRelatedObjects(instance, logger, desiredSubscription)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *StorageSystemReconciler) isVendorCsvReady(instance *odfv1alpha1.StorageSystem, logger logr.Logger) error {

	var csvName string

	if instance.Spec.Kind == VendorFlashSystemCluster() {
		csvName = IbmSubscriptionStartingCSV
	} else if instance.Spec.Kind == VendorStorageCluster() {
		csvName = OcsSubscriptionStartingCSV
	}

	csvObj := &operatorv1alpha1.ClusterServiceVersion{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name: csvName, Namespace: instance.Spec.Namespace}, csvObj)

	if err != nil {
		SetVendorCsvReadyCondition(&instance.Status.Conditions, corev1.ConditionFalse, "NotFound", err.Error())
		return err
	}

	err = r.addReferenceToRelatedObjects(instance, logger, csvObj)
	if err != nil {
		return err
	}

	if csvObj.Status.Phase == operatorv1alpha1.CSVPhaseSucceeded &&
		csvObj.Status.Reason == operatorv1alpha1.CSVReasonInstallSuccessful {

		logger.Info("Vendor csv is in ready state")
		SetVendorCsvReadyCondition(&instance.Status.Conditions, corev1.ConditionTrue, "Ready", "")
		return nil
	} else {
		err = fmt.Errorf("Vendor CSV %s is not ready", csvName)
		logger.Error(err, "Vendor csv is not ready")
		SetVendorCsvReadyCondition(&instance.Status.Conditions, corev1.ConditionFalse, "NotReady", err.Error())
		return err
	}
}

// RemoveSubscriptions deletes any managed Subscriptions that do not have any
// existing CRs that their operators reconcile.
func (r *StorageSystemReconciler) RemoveSubscriptions() error {
	var err error

	for _, kind := range KnownKinds {
		subs := GetSubscriptions(kind)
		for _, sub := range subs {
			if clientErr := r.Client.Delete(context.TODO(), sub); clientErr != nil {
				if err == nil {
					err = clientErr
				} else {
					err = fmt.Errorf("%w; ", clientErr)
				}
			}
		}
	}

	return err
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

	return []*operatorv1alpha1.Subscription{noobaaSubscription, ocsSubscription}
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
