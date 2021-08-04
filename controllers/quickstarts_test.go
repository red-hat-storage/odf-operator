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

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	consolev1 "github.com/openshift/api/console/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

var (
	cases = []struct {
		quickstartName string
	}{
		{
			quickstartName: "getting-started-odf",
		},
		{
			quickstartName: "odf-configuration",
		},
	}

	testQuickStartCRD = extv1.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name:      "consolequickstarts.console.openshift.io",
			Namespace: "",
		},
	}
)

func TestQuickStartYAMLs(t *testing.T) {
	for _, qs := range AllQuickStarts {
		cqs := consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, &cqs)
		assert.NoError(t, err)
	}
}
func TestEnsureQuickStarts(t *testing.T) {
	allExpectedQuickStarts := []consolev1.ConsoleQuickStart{}
	for _, qs := range AllQuickStarts {
		cqs := consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, &cqs)
		assert.NoError(t, err)
		allExpectedQuickStarts = append(allExpectedQuickStarts, cqs)
	}

	fakeReconciler, _ := GetFakeStorageSystemReconciler()
	err := operatorv1alpha1.AddToScheme(fakeReconciler.Scheme)
	assert.NoError(t, err)
	err = fakeReconciler.Client.Create(context.TODO(), &testQuickStartCRD)
	assert.NoError(t, err)
	err = fakeReconciler.ensureQuickStarts(fakeReconciler.Log)
	assert.NoError(t, err)
	for _, c := range cases {
		qs := consolev1.ConsoleQuickStart{}
		err = fakeReconciler.Client.Get(context.TODO(), types.NamespacedName{
			Name: c.quickstartName,
		}, &qs)
		assert.NoError(t, err)
		found := consolev1.ConsoleQuickStart{}
		expected := consolev1.ConsoleQuickStart{}
		for _, cqs := range allExpectedQuickStarts {
			if qs.Name == cqs.Name {
				found = qs
				expected = cqs
				break
			}
		}
		assert.Equal(t, expected.Name, found.Name)
		assert.Equal(t, expected.Namespace, found.Namespace)
		assert.Equal(t, expected.Spec.DurationMinutes, found.Spec.DurationMinutes)
		assert.Equal(t, expected.Spec.Introduction, found.Spec.Introduction)
		assert.Equal(t, expected.Spec.DisplayName, found.Spec.DisplayName)
	}
	assert.Equal(t, len(allExpectedQuickStarts), len(getActualQuickStarts(t, cases, fakeReconciler)))
}

func getActualQuickStarts(t *testing.T, cases []struct {
	quickstartName string
}, reconciler *StorageSystemReconciler) []consolev1.ConsoleQuickStart {
	allActualQuickStarts := []consolev1.ConsoleQuickStart{}
	for _, c := range cases {
		qs := consolev1.ConsoleQuickStart{}
		err := reconciler.Client.Get(context.TODO(), types.NamespacedName{
			Name: c.quickstartName,
		}, &qs)
		if apierrors.IsNotFound(err) {
			continue
		}
		assert.NoError(t, err)
		allActualQuickStarts = append(allActualQuickStarts, qs)
	}
	return allActualQuickStarts
}
