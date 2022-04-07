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

	"github.com/go-logr/logr"
	"go.uber.org/multierr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
)

// SubscriptionReconciler reconciles a Subscription object
type SubscriptionReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder *EventReporter
}

//+kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions/finalizers,verbs=update
//+kubebuilder:rbac:groups=operators.coreos.com,resources=installplans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SubscriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)

	instance := &operatorv1alpha1.Subscription{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Subscription instance not found.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	err = r.ensureSubscriptions(logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SubscriptionReconciler) ensureSubscriptions(logger logr.Logger) error {

	var err error

	subsList := map[odfv1alpha1.StorageKind][]*operatorv1alpha1.Subscription{}
	subsList[StorageClusterKind] = GetSubscriptions(StorageClusterKind)

	ssList := &odfv1alpha1.StorageSystemList{}
	err = r.Client.List(context.TODO(), ssList)
	if err != nil {
		return err
	}

	for _, ss := range ssList.Items {
		subsList[ss.Spec.Kind] = GetSubscriptions(ss.Spec.Kind)
	}

	for _, subs := range subsList {
		for _, sub := range subs {
			errSub := EnsureDesiredSubscription(r.Client, sub)
			if errSub != nil {
				logger.Error(errSub, "failed to ensure Subscription", "Subscription", sub.Name)
				err = fmt.Errorf("failed to ensure Subscriptions")
				multierr.AppendInto(&err, fmt.Errorf("failed to ensure Subscriptions"))
			}
		}
	}

	if err != nil {
		return err
	}

	for kind := range subsList {
		for _, csvName := range GetVendorCsvNames(kind) {
			_, csvErr := EnsureVendorCsv(r.Client, csvName)
			if csvErr != nil {
				multierr.AppendInto(&err, csvErr)
			}
		}
	}

	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {

	generationChangedPredicate := predicate.GenerationChangedPredicate{}

	predicateFunc := func(obj runtime.Object) bool {
		instance, ok := obj.(*operatorv1alpha1.Subscription)
		if !ok {
			return false
		}

		// ignore if not a odf-operator subscription
		if instance.Spec.Package != "odf-operator" {
			return false
		}

		return true
	}

	subscriptionPredicate := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return predicateFunc(e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return predicateFunc(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return predicateFunc(e.ObjectNew)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return predicateFunc(e.Object)
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Subscription{},
			builder.WithPredicates(generationChangedPredicate, subscriptionPredicate)).
		Owns(&operatorv1alpha1.Subscription{},
			builder.WithPredicates(generationChangedPredicate)).
		Complete(r)
}
