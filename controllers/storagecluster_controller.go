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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
)

const (
	HasStorageSystemAnnotation = "storagesystem.odf.openshift.io/watched-by"
)

// StorageClusterReconciler reconciles a StorageCluster object
type StorageClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder *EventReporter
}

//+kubebuilder:rbac:groups=ocs.openshift.io,resources=storageclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ocs.openshift.io,resources=storageclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ocs.openshift.io,resources=storageclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=odf.openshift.io,resources=storagesystems,verbs=get;list;create;update;patch;delete

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *StorageClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &ocsv1.StorageCluster{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("StorageCluster instance not found.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	logger.Info("StorageCluster instance found.")

	// get list of StorageSystems
	storageSystemList := &odfv1alpha1.StorageSystemList{}
	err = r.Client.List(context.TODO(), storageSystemList, &client.ListOptions{Namespace: instance.Namespace})
	if err != nil {
		logger.Error(err, "Failed to list the StorageSystem.")
		return ctrl.Result{}, err
	}

	storageSystem := filterStorageSystem(storageSystemList, instance.Name)
	if storageSystem == nil {
		logger.Info("StorageSystem not found for the StorageCluster, will create one.")

		storageSystem = &odfv1alpha1.StorageSystem{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name + "-storagesystem",
				Namespace: instance.Namespace,
			},
			Spec: odfv1alpha1.StorageSystemSpec{
				Name:      instance.Name,
				Namespace: instance.Namespace,
				Kind:      VendorStorageCluster(),
			},
		}

		// create StorageSystem for storageCluster
		err = r.Client.Create(context.TODO(), storageSystem)
		if err != nil {
			logger.Error(err, "Failed to create StorageSystem.", "StorageSystem", storageSystem.Name)
			return ctrl.Result{}, err
		}
		logger.Info("StorageSystem created for the StorageCluster.")
		r.Recorder.ReportIfNotPresent(instance, corev1.EventTypeNormal, EventReasonCreationSucceeded,
			fmt.Sprintf("StorageSystem %s created for the StorageCluster %s.", storageSystem.Name, instance.Name))
	} else {
		logger.Info("StorageSystem is already present for the StorageCluster.", "StorageSystem", storageSystem.Name)
	}

	return ctrl.Result{}, nil
}

func filterStorageSystem(storageSystemList *odfv1alpha1.StorageSystemList, storageClusterName string) *odfv1alpha1.StorageSystem {

	for _, ss := range storageSystemList.Items {
		if ss.Spec.Name == storageClusterName && ss.Spec.Kind == VendorStorageCluster() {
			return &ss
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {

	storageClusterPredicate := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&ocsv1.StorageCluster{}, builder.WithPredicates(storageClusterPredicate)).
		Complete(r)
}
