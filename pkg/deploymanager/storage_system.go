package deploymanager

import (
	"context"
	"fmt"
	"time"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	storageClusterName = "ocs-storagecluster"
	stoargeSystemName  = "ocs-storagecluster-storagesystem"
)

// GetDefaultStorageCluster returns the default StorageCluster
func GetDefaultStorageCluster() *ocsv1.StorageCluster {
	return &ocsv1.StorageCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      storageClusterName,
			Namespace: InstallNamespace,
		},
		Spec: ocsv1.StorageClusterSpec{
			// We are using this dummy version number to avoid creation of any ceph resources
			// This ensures the storagecluster is created & deleted quickly
			Version: "0.0.0.0.0",
		},
	}
}

// GetDefaultStorageSystem returns the default StorageSystem
func GetDefaultStorageSystem() *odfv1alpha1.StorageSystem {
	return &odfv1alpha1.StorageSystem{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stoargeSystemName,
			Namespace: InstallNamespace,
		},
		Spec: odfv1alpha1.StorageSystemSpec{
			Kind:      "storagecluster.ocs.openshift.io/v1",
			Name:      storageClusterName,
			Namespace: InstallNamespace,
		},
	}
}

// CreateStorageSystem creates an default StorageSystem
func (d *DeployManager) CreateStorageSystem() error {
	// Create the default StorageSystem
	storageSystem := GetDefaultStorageSystem()
	err := d.Client.Create(d.Ctx, storageSystem)
	if err != nil {
		return err
	}

	// Creating the storage cluster
	storageCluster := GetDefaultStorageCluster()
	err = d.Client.Create(d.Ctx, storageCluster)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

// CheckStorageSystemCondition verifies if the StorageSystem has the given conditions
func (d *DeployManager) CheckStorageSystemCondition() error {
	timeout := 600 * time.Second
	interval := 10 * time.Second

	lastReason := ""

	err := utilwait.PollUntilContextTimeout(d.Ctx, interval, timeout, true, func(context.Context) (done bool, err error) {
		storageSystem := &odfv1alpha1.StorageSystem{}
		err = d.Client.Get(d.Ctx, types.NamespacedName{
			Name:      stoargeSystemName,
			Namespace: InstallNamespace,
		}, storageSystem)
		if err != nil {
			lastReason = fmt.Sprintf("Failed to get StorageSystem: %v", err)
			return false, nil
		}
		conditions := storageSystem.Status.Conditions
		if len(conditions) == 0 {
			lastReason = "StorageSystem does not have any condition"
			return false, nil
		}

		// Check if the StorageSystem is having desired conditions
		vendorCsvReady := conditionsv1.IsStatusConditionPresentAndEqual(conditions, odfv1alpha1.ConditionVendorCsvReady, corev1.ConditionTrue)
		vendorSystemPresent := conditionsv1.IsStatusConditionPresentAndEqual(conditions, odfv1alpha1.ConditionVendorSystemPresent, corev1.ConditionTrue)
		available := conditionsv1.IsStatusConditionPresentAndEqual(conditions, conditionsv1.ConditionAvailable, corev1.ConditionTrue)
		progressing := conditionsv1.IsStatusConditionPresentAndEqual(conditions, conditionsv1.ConditionProgressing, corev1.ConditionFalse)
		storageSystemInvalid := conditionsv1.IsStatusConditionPresentAndEqual(conditions, odfv1alpha1.ConditionStorageSystemInvalid, corev1.ConditionFalse)

		if !available || !progressing || !storageSystemInvalid || !vendorCsvReady || !vendorSystemPresent {
			lastReason = "waiting on storagesystem to be ready"
			return false, nil
		}
		// If we reach here means storagesystem is ready now
		return true, nil
	})
	if err != nil {
		return fmt.Errorf(lastReason)
	}

	return nil
}

// CheckStorageSystemOwnerRef checks the owner reference in the storagecluster
func (d *DeployManager) CheckStorageSystemOwnerRef() error {
	storageCluster := &ocsv1.StorageCluster{}
	err := d.Client.Get(d.Ctx, types.NamespacedName{
		Name:      storageClusterName,
		Namespace: InstallNamespace},
		storageCluster)
	if err != nil {
		return err
	}

	ownerRefs := storageCluster.GetOwnerReferences()
	if len(ownerRefs) == 0 {
		return fmt.Errorf("StorageCluster does not have any owner reference")
	}

	for _, ref := range ownerRefs {
		if ref.Kind == "StorageSystem" && ref.Name == stoargeSystemName {
			return nil
		}
	}
	return fmt.Errorf("StorageCluster does not have the correct owner reference")
}

// DeleteStorageSystemAndWait deletes the default StorageSystem and waits till its completely gone
func (d *DeployManager) DeleteStorageSystemAndWait() error {
	storageSystem := &odfv1alpha1.StorageSystem{}
	err := d.Client.Get(d.Ctx, types.NamespacedName{
		Name:      stoargeSystemName,
		Namespace: InstallNamespace},
		storageSystem)
	if err != nil {
		return err
	}
	err = d.Client.Delete(d.Ctx, storageSystem)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Wait for storagesystem to be deleted
	timeout := 600 * time.Second
	interval := 10 * time.Second
	lastReason := ""

	err = utilwait.PollUntilContextTimeout(d.Ctx, interval, timeout, true, func(context.Context) (done bool, err error) {
		existingStorageSystem := &odfv1alpha1.StorageSystem{}
		err = d.Client.Get(d.Ctx, types.NamespacedName{
			Name:      stoargeSystemName,
			Namespace: InstallNamespace},
			existingStorageSystem)
		if err != nil && !errors.IsNotFound(err) {
			lastReason = "Some error in storagesystem deletion"
			return false, nil
		}
		if err == nil {
			lastReason = "waiting for storagesystem to be deleted"
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return fmt.Errorf(lastReason)
	}

	return nil
}
