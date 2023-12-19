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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
)

func TestReconcile(t *testing.T) {
	testCases := []struct {
		label                   string
		AlreadyHasStorageSystem bool
	}{
		{
			label:                   "create StorageSystem for StorageCluster if does not exist",
			AlreadyHasStorageSystem: false,
		},
		{
			label:                   "no error for StorageCluster if StorageSystem already exists",
			AlreadyHasStorageSystem: true,
		},
	}

	for i, tc := range testCases {
		t.Logf("Case %d: %s\n", i+1, tc.label)

		fakeStorageCluster := GetFakeStorageCluster()
		fakeReconciler := GetFakeStorageClusterReconciler(t, fakeStorageCluster)
		fakeStorageSystem := GetFakeStorageSystem(StorageClusterKind)
		fakeStorageSystem.Name = fakeStorageCluster.Name + "-storagesystem"
		fakeStorageSystem.Spec.Name = fakeStorageCluster.Name

		if tc.AlreadyHasStorageSystem {
			err := fakeReconciler.Client.Create(context.TODO(), fakeStorageSystem)
			assert.NoError(t, err)
		}

		_, err := fakeReconciler.Reconcile(
			context.TODO(),
			ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      fakeStorageCluster.Name,
					Namespace: fakeStorageCluster.Namespace,
				},
			},
		)
		assert.NoError(t, err)

		foundStorageSystem := &odfv1alpha1.StorageSystem{}
		err = fakeReconciler.Client.Get(context.TODO(), types.NamespacedName{
			Name: fakeStorageSystem.Name, Namespace: fakeStorageSystem.Namespace}, foundStorageSystem)
		assert.NoError(t, err)

		foundStorageCluster := &ocsv1.StorageCluster{}
		err = fakeReconciler.Client.Get(context.TODO(), types.NamespacedName{
			Name: fakeStorageCluster.Name, Namespace: fakeStorageCluster.Namespace}, foundStorageCluster)
		assert.NoError(t, err)
	}
}
