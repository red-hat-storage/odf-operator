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

	"github.com/red-hat-data-services/odf-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestReconcile(t *testing.T) {
	testCases := []struct {
		label                   string
		AlreadyHasStorageSystem bool
		expectedStorageSystem   bool
	}{
		{
			label:                   "create StorageSystem for StorageCluster if does not exist one",
			AlreadyHasStorageSystem: false,
			expectedStorageSystem:   true,
		},
		{
			label:                   "do not create StorageSystem for StorageCluster if it does exist",
			AlreadyHasStorageSystem: true,
			expectedStorageSystem:   true,
		},
	}

	for i, tc := range testCases {
		t.Logf("Case %d: %s\n", i+1, tc.label)

		fakeReconciler, fakeStorageCluster := GetFakeStorageClusterReconciler()
		_ = v1alpha1.AddToScheme(fakeReconciler.Scheme)

		fakeStorageSystem := GetFakeStorageSystem()

		if tc.AlreadyHasStorageSystem {
			err := fakeReconciler.Client.Create(context.TODO(), fakeStorageSystem)
			assert.NoError(t, err)
		}

		_, err := fakeReconciler.Reconcile(
			context.TODO(),
			ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "fake-storage-cluster",
					Namespace: "fake-namespace",
				},
			},
		)
		assert.NoError(t, err)

		if tc.AlreadyHasStorageSystem {
			err = fakeReconciler.Client.Get(context.TODO(), types.NamespacedName{Name: "fake-storage-system", Namespace: "fake-namespace"}, fakeStorageSystem)
		} else {
			err = fakeReconciler.Client.Get(context.TODO(), types.NamespacedName{Name: fakeStorageCluster.Name, Namespace: fakeStorageCluster.Namespace}, fakeStorageSystem)
		}
		assert.NoError(t, err)

		err = fakeReconciler.Client.Get(context.TODO(), types.NamespacedName{Name: fakeStorageCluster.Name, Namespace: fakeStorageCluster.Namespace}, fakeStorageCluster)
		assert.NoError(t, err)

		_, ok := fakeStorageCluster.ObjectMeta.Annotations[HasStorageSystemAnnotation]
		assert.True(t, ok)
	}
}
