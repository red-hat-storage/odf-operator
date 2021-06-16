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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
)

func GetFakeStorageSystem() *odfv1alpha1.StorageSystem {
	return &odfv1alpha1.StorageSystem{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-storage-system",
			Namespace: "fake-namespace",
		},
		Spec: odfv1alpha1.StorageSystemSpec{
			Kind:      odfv1alpha1.StorageCluster,
			Name:      "fake-storage-cluster",
			NameSpace: "fake-namespace",
		},
	}
}

func GetFakeStorageSystemReconciler() (*StorageSystemReconciler, *odfv1alpha1.StorageSystem) {

	fakeStorageSystem := GetFakeStorageSystem()

	scheme := runtime.NewScheme()
	_ = odfv1alpha1.AddToScheme(scheme)

	fakeStorageSystemReconciler := &StorageSystemReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(fakeStorageSystem).Build(),
		Log:    ctrl.Log.WithName("controllers").WithName("StorageSystem"),
		Scheme: scheme,
	}

	return fakeStorageSystemReconciler, fakeStorageSystem
}
