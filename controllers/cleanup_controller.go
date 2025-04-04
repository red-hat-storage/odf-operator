/*
Copyright 2025 Red Hat OpenShift Data Foundation.

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
	"slices"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	odfv1a1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	"github.com/red-hat-storage/odf-operator/pkg/util"
)

var (
	StorageClusterKind = odfv1a1.StorageKind("storagecluster.ocs.openshift.io/v1")
	FlashSystemKind    = odfv1a1.StorageKind("flashsystemcluster.odf.ibm.com/v1alpha1")
)

type CleanupReconciler struct {
	client.Client

	Scheme            *runtime.Scheme
	OperatorNamespace string
}

//+kubebuilder:rbac:groups=odf.openshift.io,resources=storagesystems,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=odf.openshift.io,resources=storagesystems/finalizers,verbs=update
//+kubebuilder:rbac:groups=ocs.openshift.io,resources=storageclusters,verbs=get;list;update;patch;delete
//+kubebuilder:rbac:groups=odf.ibm.com,resources=flashsystemclusters,verbs=get;list;update;patch;delete

func (r *CleanupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("starting reconcile")

	instance := &odfv1a1.StorageSystem{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); errors.IsNotFound(err) {
		logger.Info("storagesystem instance not found")
		return ctrl.Result{}, nil
	} else if err != nil {
		// Error reading the object - requeue the request.
		logger.Error(err, "error reading the object")
		return ctrl.Result{}, err
	}

	logger.Info("storagesystem instance found")

	if err := r.safelyDeleteStorageSystem(ctx, logger, instance); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconcile completed successfully")
	return ctrl.Result{}, nil
}

func (r *CleanupReconciler) safelyDeleteStorageSystem(ctx context.Context, logger logr.Logger, instance *odfv1a1.StorageSystem) error {

	cr := &metav1.PartialObjectMetadata{}
	cr.Name = instance.Spec.Name
	cr.Namespace = instance.Spec.Namespace

	switch instance.Spec.Kind {
	case StorageClusterKind:
		cr.APIVersion = "ocs.openshift.io/v1"
		cr.Kind = "StorageCluster"
	case FlashSystemKind:
		cr.APIVersion = "odf.ibm.com/v1alpha1"
		cr.Kind = "FlashSystemCluster"
	}

	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cr), cr); err != nil {
		msg := fmt.Sprintf("failed getting %s named %s", cr.Kind, cr.Name)
		logger.Error(err, msg)
		return err
	}

	// remove the storagesystem owner reference from the CR
	updatedRefs := slices.DeleteFunc(cr.OwnerReferences, func(oRef metav1.OwnerReference) bool {
		return oRef.Kind == "StorageSystem"
	})

	if len(cr.OwnerReferences) > len(updatedRefs) {
		logger.Info("removing owner reference", "kind", cr.Kind, "name", cr.Name)

		patch := client.MergeFrom(cr.DeepCopy())
		cr.OwnerReferences = updatedRefs

		if err := r.Client.Patch(ctx, cr, patch); err != nil {
			logger.Error(err, "failed removing owner references", "kind", cr.Kind, "name", cr.Name)
			return err
		}
	}

	// delete the storagesystem
	if err := r.Client.Delete(ctx, instance); err != nil {
		logger.Error(err, "failed deleting storagesystem")
		return err
	}

	// remove the finalizer
	instance.ObjectMeta.Finalizers = util.RemoveFromSlice(instance.ObjectMeta.Finalizers, "storagesystem.odf.openshift.io")

	if err := r.Client.Update(context.TODO(), instance); err != nil {
		logger.Error(err, "failed deleting storagesystem")
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CleanupReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&odfv1a1.StorageSystem{},
			builder.WithPredicates(createOnlyPredicate),
		).
		Complete(r)
}
