/*
Copyright 2026 Red Hat OpenShift Data Foundation.
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
	"sync/atomic"
	"testing"

	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func newTestResources() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
	}
}

func newOdfSubscription(ns string) *opv1a1.Subscription {
	return &opv1a1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      OdfSubscriptionPackage,
			Namespace: ns,
		},
		Spec: &opv1a1.SubscriptionSpec{
			Package:                OdfSubscriptionPackage,
			CatalogSource:          "test-catalog",
			CatalogSourceNamespace: "openshift-marketplace",
			Config: &opv1a1.SubscriptionConfig{
				Tolerations: []corev1.Toleration{
					{
						Key:      "node.ocs.openshift.io/storage",
						Operator: "Equal",
						Value:    "true",
						Effect:   "NoSchedule",
					},
				},
			},
		},
	}
}

func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(opv1a1.AddToScheme(s))
	utilruntime.Must(corev1.AddToScheme(s))
	return s
}

// TestEnsureDesiredSubscription_ResourcesPreserved verifies that user-set
// resource requests/limits on a subscription's Config are not wiped during
// reconciliation under normal (no-race) conditions.
func TestEnsureDesiredSubscription_ResourcesPreserved(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme()

	const (
		operatorNs = "openshift-storage"
		targetNs   = "ibm-spectrum-scale"
		cnsaPkg    = "cnsa-dependencies"
	)

	userResources := newTestResources()

	cnsaSub := &opv1a1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cnsaPkg,
			Namespace: targetNs,
		},
		Spec: &opv1a1.SubscriptionSpec{
			Package: cnsaPkg,
			Channel: "stable-4.18",
			Config: &opv1a1.SubscriptionConfig{
				Resources: userResources,
			},
		},
	}

	cli := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(newOdfSubscription(operatorNs), cnsaSub).
		Build()

	record := &OlmPkgRecord{
		Channel:   "stable-4.19",
		Csv:       cnsaPkg + ".v4.19.0",
		Pkg:       cnsaPkg,
		Namespace: targetNs,
	}

	if err := EnsureDesiredSubscription(context.Background(), cli, record, providerNameIBM); err != nil {
		t.Fatalf("EnsureDesiredSubscription() error: %v", err)
	}

	result := &opv1a1.Subscription{}
	if err := cli.Get(context.Background(), client.ObjectKeyFromObject(cnsaSub), result); err != nil {
		t.Fatalf("failed to get subscription after reconcile: %v", err)
	}

	if result.Spec.Config == nil || result.Spec.Config.Resources == nil {
		t.Fatal("Resources were wiped: Spec.Config.Resources is nil after reconciliation")
	}

	gotCPUReq := result.Spec.Config.Resources.Requests[corev1.ResourceCPU]
	wantCPUReq := userResources.Requests[corev1.ResourceCPU]
	if !gotCPUReq.Equal(wantCPUReq) {
		t.Errorf("CPU request = %s, want %s", gotCPUReq.String(), wantCPUReq.String())
	}

	gotMemLimit := result.Spec.Config.Resources.Limits[corev1.ResourceMemory]
	wantMemLimit := userResources.Limits[corev1.ResourceMemory]
	if !gotMemLimit.Equal(wantMemLimit) {
		t.Errorf("Memory limit = %s, want %s", gotMemLimit.String(), wantMemLimit.String())
	}

	if result.Spec.Channel != "stable-4.19" {
		t.Errorf("Channel = %s, want stable-4.19", result.Spec.Channel)
	}
}

// TestEnsureDesiredSubscription_CacheRaceWipesResources reproduces the race
// condition where the informer cache updates between the List (in
// GetDesiredSubscription) and the Get (inside CreateOrUpdate).
//
// Timeline:
//  1. List returns the subscription WITHOUT the user's Resources (stale cache).
//  2. The cache updates mid-reconcile.
//  3. CreateOrUpdate's internal Get returns the subscription WITH Resources (fresh).
//  4. The mutate function does sub.Spec = desiredSubscription.Spec (stale, no Resources).
//  5. Update succeeds because ResourceVersion from the fresh Get is valid.
//  6. Resources are wiped.
//
// We use two clients: an intercepted one for the reconcile (to simulate the
// race) and the bare fake client for final verification (to read what was
// actually persisted, without the interceptor injecting Resources).
func TestEnsureDesiredSubscription_CacheRaceWipesResources(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme()

	const (
		operatorNs = "openshift-storage"
		targetNs   = "ibm-spectrum-scale"
		cnsaPkg    = "cnsa-dependencies"
	)

	userResources := newTestResources()

	// The subscription in the fake store does NOT have Resources.
	// This represents the stale cache state at the time of the List call.
	cnsaSubStale := &opv1a1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cnsaPkg,
			Namespace: targetNs,
		},
		Spec: &opv1a1.SubscriptionSpec{
			Package: cnsaPkg,
			Channel: "stable-4.18",
			Config: &opv1a1.SubscriptionConfig{
				Tolerations: []corev1.Toleration{
					{
						Key:      "node.ocs.openshift.io/storage",
						Operator: "Equal",
						Value:    "true",
						Effect:   "NoSchedule",
					},
				},
			},
		},
	}

	baseCli := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(newOdfSubscription(operatorNs), cnsaSubStale).
		Build()

	// listDone is set after the List for the target namespace completes.
	// injected is set after the first Get injection, so only one Get
	// (the one inside CreateOrUpdate) sees the "fresh" Resources.
	var listDone atomic.Bool
	var injected atomic.Bool

	interceptedCli := interceptor.NewClient(baseCli, interceptor.Funcs{
		List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
			if err := c.List(ctx, list, opts...); err != nil {
				return err
			}
			if subList, ok := list.(*opv1a1.SubscriptionList); ok {
				for i := range subList.Items {
					if subList.Items[i].Namespace == targetNs {
						listDone.Store(true)
					}
				}
			}
			return nil
		},
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if err := c.Get(ctx, key, obj, opts...); err != nil {
				return err
			}
			// Simulate cache refresh: only on the first Get for the CNSA
			// subscription after the List has completed (this is the Get
			// inside CreateOrUpdate).
			sub, ok := obj.(*opv1a1.Subscription)
			if ok && sub.Name == cnsaPkg && sub.Namespace == targetNs &&
				listDone.Load() && injected.CompareAndSwap(false, true) {
				if sub.Spec.Config == nil {
					sub.Spec.Config = &opv1a1.SubscriptionConfig{}
				}
				sub.Spec.Config.Resources = userResources.DeepCopy()
			}
			return nil
		},
	})

	record := &OlmPkgRecord{
		Channel:   "stable-4.19",
		Csv:       cnsaPkg + ".v4.19.0",
		Pkg:       cnsaPkg,
		Namespace: targetNs,
	}

	if err := EnsureDesiredSubscription(context.Background(), interceptedCli, record, providerNameIBM); err != nil {
		t.Fatalf("EnsureDesiredSubscription() error: %v", err)
	}

	// Read persisted state from the base client (no interceptor) to see
	// what was actually written to the store.
	result := &opv1a1.Subscription{}
	if err := baseCli.Get(context.Background(), client.ObjectKeyFromObject(cnsaSubStale), result); err != nil {
		t.Fatalf("failed to get subscription after reconcile: %v", err)
	}

	// With the current code, the full-spec replacement in the mutate
	// function overwrites the fresh Get data (which had Resources) with
	// the stale List data (which did not). This confirms the race.
	if result.Spec.Config == nil || result.Spec.Config.Resources == nil {
		t.Fatal("user-set Resources were wiped due to cache race between List and Get. " +
			"The stale List data overwrote the fresh Get data via `sub.Spec = desiredSubscription.Spec`.")
	}

	gotCPUReq := result.Spec.Config.Resources.Requests[corev1.ResourceCPU]
	wantCPUReq := userResources.Requests[corev1.ResourceCPU]
	if !gotCPUReq.Equal(wantCPUReq) {
		t.Errorf("CPU request = %s, want %s", gotCPUReq.String(), wantCPUReq.String())
	}
}
