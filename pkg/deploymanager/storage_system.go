package deploymanager

import (
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDefaultStorageSystem returns the default storagesystem
func GetDefaultStorageSystem() *odfv1alpha1.StorageSystem {

	return &odfv1alpha1.StorageSystem{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ocs-storagecluster-storagesystem",
			Namespace: InstallNamespace,
		},
		Spec: odfv1alpha1.StorageSystemSpec{
			Kind:      "storagecluster.ocs.openshift.io/v1",
			Name:      "ocs-storagecluster",
			Namespace: InstallNamespace,
		},
	}
}

// CreateStorageSystem creates an default StorageSystem
func (d *DeployManager) CreateStorageSystem() error {
	storageSystem := GetDefaultStorageSystem()
	return d.Client.Create(d.Ctx, storageSystem)
}

// DeleteStorageSystem deletes an default StorageSystem
func (d *DeployManager) DeleteStorageSystem() error {
	storageSystem := GetDefaultStorageSystem()
	return d.Client.Delete(d.Ctx, storageSystem)
}
