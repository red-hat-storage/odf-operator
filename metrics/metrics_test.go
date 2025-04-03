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

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var defaultRegistry = metrics.Registry
var find string = "odf_system_map"

func TestReportODFSystemMapMetrics(t *testing.T) {
	type args struct {
		storageSystem, name, namespace, kind, apiVersion string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "StorageSystem to StorageCluster",
			args: args{
				storageSystem: "StorageSystem1",
				name:          "StorageCluster1",
				namespace:     "Namespace1",
				kind:          "StorageCluster",
				apiVersion:    "ocs.openshift.io/v1",
			},
		},
		{
			name: "StorageSystem to Flashsystem",
			args: args{
				storageSystem: "StorageSystem2",
				name:          "FlashSystem1",
				namespace:     "Namespace1",
				kind:          "Flashsystem",
				apiVersion:    "odf.ibm.com/v1alpha1",
			},
		},
	}
	for n, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReportODFSystemMapMetrics(tt.args.storageSystem, tt.args.name, tt.args.namespace, tt.args.kind, tt.args.apiVersion)
			count, err := testutil.GatherAndCount(defaultRegistry, find)
			assert.NoError(t, err)
			assert.Equal(t, n+1, count)
			problems, err := testutil.GatherAndLint(defaultRegistry, find)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(problems))
		})
	}
}
