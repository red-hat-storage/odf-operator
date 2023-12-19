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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ibmv1alpha1 "github.com/IBM/ibm-storage-odf-operator/api/v1alpha1"
	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
)

func (r *StorageSystemReconciler) deleteResources(instance *odfv1alpha1.StorageSystem, logger logr.Logger) error {

	var backendStorage client.Object

	r.deleteQuickStarts(logger, instance)
	if instance.Spec.Kind == VendorStorageCluster() {
		backendStorage = &ocsv1.StorageCluster{}
	} else if instance.Spec.Kind == VendorFlashSystemCluster() {
		backendStorage = &ibmv1alpha1.FlashSystemCluster{}
	}

	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name: instance.Spec.Name, Namespace: instance.Spec.Namespace},
		backendStorage)

	if errors.IsNotFound(err) {
		logger.Info("Deleted successfully", "Kind", instance.Spec.Kind, "Name", instance.Spec.Name)
		return nil
	} else if err != nil {
		logger.Error(err, "Failed to Get", "Kind", instance.Spec.Kind, "Name", instance.Spec.Name)
		return err
	}

	err = r.Client.Delete(context.TODO(), backendStorage)
	if err != nil {
		logger.Error(err, "Failed to delete", "Kind", instance.Spec.Kind, "Name", instance.Spec.Name)
		return err
	}

	logger.Info("Waiting for deletion", "Kind", instance.Spec.Kind, "Name", instance.Spec.Name)

	return fmt.Errorf("Waiting for %s %s to be deleted", instance.Spec.Kind, instance.Spec.Name)
}

func (r *StorageSystemReconciler) areAllStorageSystemsMarkedForDeletion(namespace string) (bool, error) {

	var storageSystems odfv1alpha1.StorageSystemList
	err := r.Client.List(context.TODO(), &storageSystems, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return false, err
	}

	var deleteCount int = 0
	for _, ss := range storageSystems.Items {
		if !ss.DeletionTimestamp.IsZero() {
			deleteCount++
		}
	}
	if len(storageSystems.Items) == deleteCount {
		return true, nil
	}

	return false, nil
}
