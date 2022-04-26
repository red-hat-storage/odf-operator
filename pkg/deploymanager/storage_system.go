package deploymanager

import (
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

// GetDefaultStorageCluster returns the default StorageCluster
func GetDefaultStorageCluster() *ocsv1.StorageCluster {
	return &ocsv1.StorageCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ocs-storagecluster",
			Namespace: InstallNamespace,
		},
		Spec: ocsv1.StorageClusterSpec{
			Version: "0.0.0.0.0",
		},
	}
}

// CreateStorageSystem creates an default StorageSystem
func (d *DeployManager) CreateStorageSystem() error {
	// Creating a required storage cluster
	storageCluster := GetDefaultStorageCluster()
	err := d.Client.Create(d.Ctx, storageCluster)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	err = d.Client.Get(d.Ctx, types.NamespacedName{
		Name:      "ocs-storagecluster",
		Namespace: InstallNamespace}, storageCluster)
	if err != nil {
		return err
	}

	// Checking if the storage system has been automatically created
	storageSystem := &odfv1alpha1.StorageSystem{}
	err = d.Client.Get(d.Ctx, types.NamespacedName{
		Name:      "ocs-storagecluster-storagesystem",
		Namespace: InstallNamespace}, storageSystem)
	if err != nil {
		return err
	}
	return nil
}

// CheckStorageSystemCondition verifies if the StorageSystem has the given conditions
func (d *DeployManager) CheckStorageSystemCondition() error {
	timeout := 600 * time.Second
	interval := 10 * time.Second

	lastReason := ""

	err := utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		storageSystem := &odfv1alpha1.StorageSystem{}
		err = d.Client.Get(d.Ctx, types.NamespacedName{
			Name:      "ocs-storagecluster-storagesystem",
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

//  CheckStorageSystemOwnerRef cheks the owner reference in the storagecluster
func (d *DeployManager) CheckStorageSystemOwnerRef() error {
	storageCluster := &ocsv1.StorageCluster{}
	err := d.Client.Get(d.Ctx, types.NamespacedName{
		Name:      "ocs-storagecluster",
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
		if ref.Kind == "StorageSystem" && ref.Name == "ocs-storagecluster-storagesystem" {
			return nil
		}
	}
	return fmt.Errorf("StorageCluster does not have the correct owner reference")
}

// DeleteStorageSystem deletes the default StorageSystem and waits till its completely gone
func (d *DeployManager) DeleteStorageSystemAndWait() error {
	storageSystem := &odfv1alpha1.StorageSystem{}
	err := d.Client.Get(d.Ctx, types.NamespacedName{
		Name:      "ocs-storagecluster-storagesystem",
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

	err = utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		existingStorageSystem:= &odfv1alpha1.StorageSystem{}
		err = d.Client.Get(d.Ctx, types.NamespacedName{
				Name:    "ocs-storagecluster-storagesystem",
				Namespace: InstallNamespace},
				existingStorageSystem)
		if err != nil && !errors.IsNotFound(err) {
			return false, nil
		}
		if err == nil {
			d.Log.Info("Waiting on storagesystem to be deleted")
			return false, nil
		}
		return true, nil
	})

	if err != nil{
		return err
	}

	return nil
}
