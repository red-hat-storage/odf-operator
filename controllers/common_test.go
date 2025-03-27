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

	"github.com/go-logr/logr"
	consolev1 "github.com/openshift/api/console/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	testScheme  *runtime.Scheme
	testContext context.Context
	testClient  client.Client
	testLogger  logr.Logger
)

func init() {

	log.SetLogger(zap.New(zap.UseDevMode(true)))

	testScheme = runtime.NewScheme()
	testContext = context.TODO()
	testClient = fake.NewClientBuilder().WithScheme(testScheme).WithRuntimeObjects().Build()
	testLogger = log.Log.WithName("test-logger")

	utilruntime.Must(consolev1.AddToScheme(testScheme))
}
