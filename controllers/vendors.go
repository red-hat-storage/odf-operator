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
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ibmv1alpha1 "github.com/IBM/ibm-storage-odf-operator/api/v1alpha1"
	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
)

var StorageClusterKind = odfv1alpha1.StorageKind(strings.ToLower(reflect.TypeOf(ocsv1.StorageCluster{}).Name()) +
	"." + ocsv1.GroupVersion.String())
var FlashSystemKind = odfv1alpha1.StorageKind(strings.ToLower(reflect.TypeOf(ibmv1alpha1.FlashSystemCluster{}).Name()) +
	"." + ibmv1alpha1.GroupVersion.String())

var KnownKinds = []odfv1alpha1.StorageKind{
	StorageClusterKind,
	FlashSystemKind,
}

// VendorStorageCluster returns GroupVersionKind
func VendorStorageCluster() odfv1alpha1.StorageKind {

	// storagecluster.ocs.openshift.io/v1
	return StorageClusterKind
}

// VendorFlashSystemCluster returns GroupVersionKind
func VendorFlashSystemCluster() odfv1alpha1.StorageKind {

	// flashsystemcluster.odf.ibm.com/v1alpha1
	return FlashSystemKind
}

func (r *StorageSystemReconciler) isVendorSystemPresent(instance *odfv1alpha1.StorageSystem, logger logr.Logger) error {

	var vendorSystem client.Object

	if instance.Spec.Kind == VendorStorageCluster() {
		logger.Info("get storageCluster")
		vendorSystem = &ocsv1.StorageCluster{}
	} else if instance.Spec.Kind == VendorFlashSystemCluster() {
		logger.Info("get flashSystemCluster")
		vendorSystem = &ibmv1alpha1.FlashSystemCluster{}
	}

	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Name, Namespace: instance.Spec.Namespace}, vendorSystem)
	if err == nil {
		logger.Info("Vendor system found", "Name", instance.Spec.Name)
		SetVendorSystemPresentCondition(&instance.Status.Conditions, corev1.ConditionTrue, "Found", "")
		_, err = controllerutil.CreateOrUpdate(context.TODO(), r.Client, vendorSystem, func() error {
			return controllerutil.SetOwnerReference(instance, vendorSystem, r.Scheme)
		})
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	} else if errors.IsNotFound(err) {
		logger.Error(err, "Vendor system not found", "Name", instance.Spec.Name)
		SetVendorSystemPresentCondition(&instance.Status.Conditions, corev1.ConditionFalse, "NotFound", err.Error())
	} else {
		logger.Error(err, "Vendor system fetch error", "Name", instance.Spec.Name)
		SetVendorSystemPresentCondition(&instance.Status.Conditions, corev1.ConditionUnknown, "Error", err.Error())
	}

	return err
}
