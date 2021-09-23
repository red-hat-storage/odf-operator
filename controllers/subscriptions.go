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

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
)

func FilterSubscriptionWithPackage(subs *operatorv1alpha1.SubscriptionList, pkg string) *operatorv1alpha1.Subscription {

	for _, s := range subs.Items {
		if s.Spec.Package == pkg {
			return &s
		}
	}

	return nil
}

func (r *StorageSystemReconciler) ensureSubscription(instance *odfv1alpha1.StorageSystem, logger logr.Logger) error {

	var desiredSubscription *operatorv1alpha1.Subscription

	if instance.Spec.Kind == VendorStorageCluster() {
		for _, subscription := range GetStorageClusterSubscriptions() {
			err := r.addReferenceToRelatedObjects(instance, logger, subscription)
			if err != nil {
				return err
			}
		}
		// No need to create subscription
		return nil
	} else if instance.Spec.Kind == VendorFlashSystemCluster() {
		desiredSubscription = GetFlashSystemClusterSubscriptions()[0]
		desiredSubscription.OwnerReferences = []metav1.OwnerReference{{
			APIVersion: instance.APIVersion,
			Kind:       instance.Kind,
			Name:       instance.Name,
			UID:        instance.UID,
			Controller: func() *bool { flag := true; return &flag }(),
		}}

		err := r.addReferenceToRelatedObjects(instance, logger, desiredSubscription)
		if err != nil {
			return err
		}
	}

	// create/update subscription
	existingSubscriptions := &operatorv1alpha1.SubscriptionList{}
	err := r.Client.List(context.TODO(), existingSubscriptions)
	if err == nil {
		existingSubscription := FilterSubscriptionWithPackage(existingSubscriptions, desiredSubscription.Spec.Package)
		if existingSubscription == nil {
			logger.Info("subscription not found create it")
			err = r.Client.Create(context.TODO(), desiredSubscription)
			if err != nil {
				logger.Error(err, "failed to create subscription")
				return err
			}
		} else {
			logger.Info("subscription found update it")
			desiredSubscription.Spec.DeepCopyInto(existingSubscription.Spec)
			err = r.Client.Update(context.TODO(), existingSubscription)
			if err != nil {
				logger.Error(err, "failed to update subscription")
				return err
			}
		}
	} else if errors.IsNotFound(err) {
		logger.Info("subscription not found create it")
		err = r.Client.Create(context.TODO(), desiredSubscription)
		if err != nil {
			logger.Error(err, "failed to create subscription")
			return err
		}
	} else {
		logger.Error(err, "failed to fetch subscription")
		return err
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
