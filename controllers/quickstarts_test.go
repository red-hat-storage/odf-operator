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
	consolev1 "github.com/openshift/api/console/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestQuickStartYamls(t *testing.T) {
	for _, qs := range AllQuickStarts {
		cqs := consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, &cqs)
		assert.NoError(t, err)
	}
}

func TestEnsureQuickStarts(t *testing.T) {

	scheme := createFakeScheme(t)
	ctx := context.TODO()
	cli := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects().Build()
	logger := log.Log.WithName("test-logger")

	err := ensureQuickStarts(ctx, cli, logger)
	assert.NoError(t, err)

	// Check if all quickstarts are created
	for _, qs := range AllQuickStarts {
		expected := &consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, expected)
		assert.NoError(t, err)

		found := &consolev1.ConsoleQuickStart{}
		err = cli.Get(ctx, client.ObjectKeyFromObject(expected), found)
		assert.NoError(t, err)

		assert.Equal(t, expected.Name, found.Name)
		assert.Equal(t, expected.Namespace, found.Namespace)
		assert.Equal(t, expected.Spec.DurationMinutes, found.Spec.DurationMinutes)
		assert.Equal(t, expected.Spec.Introduction, found.Spec.Introduction)
		assert.Equal(t, expected.Spec.DisplayName, found.Spec.DisplayName)
	}
}

func TestDeleteQuickStarts(t *testing.T) {

	scheme := createFakeScheme(t)
	ctx := context.TODO()
	cli := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects().Build()
	logger := log.Log.WithName("test-logger")

	deleteQuickStarts(ctx, cli, logger)

	// Check if all quickstarts are deleted
	for _, qs := range AllQuickStarts {
		expected := &consolev1.ConsoleQuickStart{}
		err := yaml.Unmarshal(qs, expected)
		assert.NoError(t, err)

		found := &consolev1.ConsoleQuickStart{}
		err = cli.Get(ctx, client.ObjectKeyFromObject(expected), found)
		assert.Error(t, err)
		assert.True(t, errors.IsNotFound(err))
	}
}
