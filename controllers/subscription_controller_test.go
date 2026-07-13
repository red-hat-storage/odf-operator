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

	"github.com/blang/semver/v4"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/operator-framework/api/pkg/lib/version"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsODFAheadOfOCP(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	utilruntime.Must(configv1.AddToScheme(scheme))
	utilruntime.Must(opv1a1.AddToScheme(scheme))

	const operatorNamespace = "openshift-storage"

	tests := []struct {
		name                  string
		odfVersion            string
		ocpVersion            string
		operatorConditionName string
		wantAhead             bool
	}{
		{
			name:                  "ODF 4.23 on OCP 4.23 is not ahead",
			odfVersion:            "4.23.0",
			ocpVersion:            "4.23.0",
			operatorConditionName: "odf-operator.v4.23.0",
			wantAhead:             false,
		},
		{
			name:                  "ODF 4.23.5 on OCP 4.23.0 is not ahead",
			odfVersion:            "4.23.5",
			ocpVersion:            "4.23.0",
			operatorConditionName: "odf-operator.v4.23.5",
			wantAhead:             false,
		},
		{
			name:                  "ODF 4.23 on OCP 5.0 is not ahead",
			odfVersion:            "4.23.0",
			ocpVersion:            "5.0.0",
			operatorConditionName: "odf-operator.v4.23.0",
			wantAhead:             false,
		},
		{
			name:                  "ODF 4.23 on OCP 5.1 is not ahead",
			odfVersion:            "4.23.0",
			ocpVersion:            "5.1.0",
			operatorConditionName: "odf-operator.v4.23.0",
			wantAhead:             false,
		},
		{
			name:                  "ODF 4.23 on OCP 4.22 is ahead",
			odfVersion:            "4.23.0",
			ocpVersion:            "4.22.0",
			operatorConditionName: "odf-operator.v4.23.0",
			wantAhead:             true,
		},
		{
			name:                  "ODF 5.0 on OCP 5.0 is not ahead",
			odfVersion:            "5.0.0",
			ocpVersion:            "5.0.0",
			operatorConditionName: "odf-operator.v5.0.0",
			wantAhead:             false,
		},
		{
			name:                  "ODF 5.0 on OCP 4.23 is ahead",
			odfVersion:            "5.0.0",
			ocpVersion:            "4.23.0",
			operatorConditionName: "odf-operator.v5.0.0",
			wantAhead:             true,
		},
		{
			name:                  "ODF 5.0 on OCP 5.1 is not ahead",
			odfVersion:            "5.0.0",
			ocpVersion:            "5.1.0",
			operatorConditionName: "odf-operator.v5.0.0",
			wantAhead:             false,
		},
		{
			name:                  "ODF 5.1 on OCP 5.0 is ahead",
			odfVersion:            "5.1.0",
			ocpVersion:            "5.0.0",
			operatorConditionName: "odf-operator.v5.1.0",
			wantAhead:             true,
		},
		{
			name:                  "ODF 5.1 on OCP 5.1 is not ahead",
			odfVersion:            "5.1.0",
			ocpVersion:            "5.1.0",
			operatorConditionName: "odf-operator.v5.1.0",
			wantAhead:             false,
		},
		{
			name:                  "ODF 5.1 on OCP 4.23 is ahead",
			odfVersion:            "5.1.0",
			ocpVersion:            "4.23.0",
			operatorConditionName: "odf-operator.v5.1.0",
			wantAhead:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			clusterVersion := &configv1.ClusterVersion{
				ObjectMeta: metav1.ObjectMeta{Name: "version"},
				Status: configv1.ClusterVersionStatus{
					History: []configv1.UpdateHistory{
						{
							State:   configv1.CompletedUpdate,
							Version: tt.ocpVersion,
						},
					},
				},
			}
			csv := &opv1a1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.operatorConditionName,
					Namespace: operatorNamespace,
				},
				Spec: opv1a1.ClusterServiceVersionSpec{
					Version: version.OperatorVersion{
						Version: semver.MustParse(tt.odfVersion),
					},
				},
			}

			cl := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(clusterVersion, csv).
				Build()

			reconciler := &SubscriptionReconciler{
				Client:                cl,
				OperatorNamespace:     operatorNamespace,
				operatorConditionName: tt.operatorConditionName,
			}

			gotAhead, err := reconciler.isODFAheadOfOCP(context.Background())
			if err != nil {
				t.Fatalf("isODFAheadOfOCP() returned error: %v", err)
			}
			if gotAhead != tt.wantAhead {
				t.Fatalf("isODFAheadOfOCP() = %v, want %v", gotAhead, tt.wantAhead)
			}
		})
	}
}
