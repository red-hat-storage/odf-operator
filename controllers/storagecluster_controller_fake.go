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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"
)

func GetFakeStorageCluster() *ocsv1.StorageCluster {
	return &ocsv1.StorageCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-vendor-system",
			Namespace: OperatorNamespace,
		},
		Spec: ocsv1.StorageClusterSpec{},
	}
}

func GetFakeStorageClusterReconciler(t *testing.T, objs ...runtime.Object) *StorageClusterReconciler {

	scheme := createFakeScheme(t)
	fakeStorageClusterReconciler := &StorageClusterReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build(),
		Scheme:   scheme,
		Recorder: NewEventReporter(record.NewFakeRecorder(1024)),
	}

	return fakeStorageClusterReconciler
}
