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
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	ibmv1alpha1 "github.com/IBM/ibm-storage-odf-operator/api/v1alpha1"
	consolev1 "github.com/openshift/api/console/v1"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
)

func GetFakeStorageSystem() *odfv1alpha1.StorageSystem {
	return &odfv1alpha1.StorageSystem{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-storage-system",
			Namespace: "fake-namespace",
		},
		Spec: odfv1alpha1.StorageSystemSpec{
			Kind:      VendorStorageCluster(),
			Name:      "fake-storage-cluster",
			Namespace: "fake-namespace",
		},
	}
}

func GetFakeStorageSystemReconciler() (*StorageSystemReconciler, *odfv1alpha1.StorageSystem) {

	fakeStorageSystem := GetFakeStorageSystem()

	scheme := runtime.NewScheme()
	_ = odfv1alpha1.AddToScheme(scheme)

	_ = consolev1.AddToScheme(scheme)

	_ = extv1.AddToScheme(scheme)

	fakeStorageSystemReconciler := &StorageSystemReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(fakeStorageSystem).Build(),
		Log:      ctrl.Log.WithName("controllers").WithName("StorageSystem"),
		Scheme:   scheme,
		Recorder: NewEventReporter(record.NewFakeRecorder(1024)),
	}

	return fakeStorageSystemReconciler, fakeStorageSystem
}

func GetFakeFlashSystemCluster() *ibmv1alpha1.FlashSystemCluster {
	return &ibmv1alpha1.FlashSystemCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-flash-system-cluster",
			Namespace: "fake-namespace",
		},
		Spec: ibmv1alpha1.FlashSystemClusterSpec{},
	}
}
