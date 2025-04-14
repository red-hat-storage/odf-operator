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
	"strings"

	"github.com/go-logr/logr"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
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

	// TODO: this check should be removed in 4.20
	// During an ODF operator upgrade, a race condition occurs where:
	//  odf-operator upgrades first to 4.19 and removes the StorageSystem ownerReference from StorageCluster
	//  ocs-operator (still at v4.18) fails to find this ownerReference, marking itself as not upgradable
	// This creates a deadlock because:
	//  ocs-operator requires the StorageSystem ownerReference to mark itself upgradable
	//  odf-operator has already deleted both the ownerReference and StorageSystem CR
	// The function ensures that ocs-operator CSV upgrades before running the cleanup, preventing this race condition.
	// In 4.19 we are not using StorageSystem CR, so the check should be removed in 4.20
	if err := r.isOcsOperatorSubAndCSVAt419(ctx, logger); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.safelyDeleteStorageSystem(ctx, logger, instance); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconcile completed successfully")
	return ctrl.Result{}, nil
}

func (r *CleanupReconciler) isOcsOperatorSubAndCSVAt419(ctx context.Context, logger logr.Logger) error {
	subscriptionList := &opv1a1.SubscriptionList{}
	if err := r.List(ctx, subscriptionList, client.InNamespace(r.OperatorNamespace)); err != nil {
		return fmt.Errorf("failed to list subscriptions: %v", err)
	}
	ocsOperatorSubscription := util.Find(subscriptionList.Items, func(sub *opv1a1.Subscription) bool {
		return sub.Spec.Package == "ocs-operator"
	})
	if ocsOperatorSubscription == nil {
		return fmt.Errorf("no subscription exists with package 'ocs-operator'")
	} else if strings.HasSuffix(ocsOperatorSubscription.Spec.Channel, "4.18") {
		return fmt.Errorf("subscription of 'ocs-operator' still points to '4.18'")
	} else if !strings.HasSuffix(ocsOperatorSubscription.Spec.Channel, "4.19") {
		// a guard that this code should be skipped in 4.19+ even if the code isn't removed
		return nil
	}
	if !strings.Contains(ocsOperatorSubscription.Status.InstalledCSV, "4.19") {
		return fmt.Errorf("waiting for 'ocs-operator' installed CSV at '4.19'")
	}

	return r.isOcsCsvReady(ctx, logger, ocsOperatorSubscription.Status.InstalledCSV)
}

func (r *CleanupReconciler) isOcsCsvReady(ctx context.Context, logger logr.Logger, ocsCsvName string) error {

	ocsCsv := &opv1a1.ClusterServiceVersion{}
	ocsCsv.Name = ocsCsvName
	ocsCsv.Namespace = OperatorNamespace

	err := r.Client.Get(ctx, client.ObjectKeyFromObject(ocsCsv), ocsCsv)
	if err != nil {
		logger.Error(err, "failed getting ocs csv", "csvName", ocsCsv)
		return err
	}

	if ocsCsv.Status.Phase != opv1a1.CSVPhaseSucceeded {
		err = fmt.Errorf("csv %s is not in succeeded state", ocsCsvName)
		logger.Error(err, "waiting for csv to be in succeeded state")
		return err
	}

	return nil
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
	instance.ObjectMeta.Finalizers = slices.DeleteFunc(instance.ObjectMeta.Finalizers, func(s string) bool {
		return s == "storagesystem.odf.openshift.io"
	})

	if err := r.Client.Update(ctx, instance); err != nil {
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
