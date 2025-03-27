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
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	opv2 "github.com/operator-framework/api/pkg/operators/v2"
	"github.com/operator-framework/operator-lib/conditions"
	"go.uber.org/multierr"
	admrv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/red-hat-storage/odf-operator/pkg/util"
)

type OlmPkgRecord struct {
	/* example
	   channel: alpha
	   csv: ocs-operator.v4.18.0
	   pkg: ocs-operator
	*/

	Channel string `yaml:"channel"`
	Csv     string `yaml:"csv"`
	Pkg     string `yaml:"pkg"`
}

type SubscriptionReconciler struct {
	client.Client

	Scheme            *runtime.Scheme
	OperatorNamespace string

	operatorConditionName string
	operatorCondition     conditions.Condition
}

//+kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions/finalizers,verbs=update
//+kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions/finalizers,verbs=update
//+kubebuilder:rbac:groups=operators.coreos.com,resources=installplans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operators.coreos.com,resources=operatorconditions,verbs=get;list;watch
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update

func (r *SubscriptionReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)

	olmPkgRecords := getOlmPkgRecord()

	if err := r.setOperatorCondition(logger, olmPkgRecords); err != nil {
		return ctrl.Result{}, err
	}

	if err := reconcileCsvWebhook(ctx, r.Client, logger, r.OperatorNamespace); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.ensureSubscriptions(logger, olmPkgRecords); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SubscriptionReconciler) ensureSubscriptions(logger logr.Logger, olmPkgRecords []*OlmPkgRecord) error {

	var combinedErr error

	for _, olmPkgRecord := range olmPkgRecords {
		if err := EnsureDesiredSubscription(r.Client, olmPkgRecord); err != nil {
			logger.Error(err, "failed to ensure subscription", "package", olmPkgRecord.Pkg)
			multierr.AppendInto(&combinedErr, err)
		}
	}

	if combinedErr != nil {
		return combinedErr
	}

	// Ensure CSVs are checked only after updating the channel of all subscriptions.
	// Checking the CSV of a single updated subscription is incorrect,
	// as there won't be any desired CSVs until all subscriptions are updated.

	for _, olmPkgRecord := range olmPkgRecords {
		if err := EnsureCsv(r.Client, olmPkgRecord); err != nil {
			multierr.AppendInto(&combinedErr, err)
		}
	}

	return combinedErr
}

func (r *SubscriptionReconciler) setOperatorCondition(logger logr.Logger, olmPkgRecords []*OlmPkgRecord) error {
	ocdList := &opv2.OperatorConditionList{}
	err := r.Client.List(context.TODO(), ocdList, client.InNamespace(r.OperatorNamespace))
	if err != nil {
		logger.Error(err, "failed to list OperatorConditions")
		return err
	}

	condMap := make(map[string]struct{}, len(olmPkgRecords))
	for _, olmPkgRecord := range olmPkgRecords {
		condMap[olmPkgRecord.Csv] = struct{}{}
	}

	for ocdIdx := range ocdList.Items {
		// skip operatorconditions of not dependent operators
		if _, exist := condMap[ocdList.Items[ocdIdx].GetName()]; !exist {
			continue
		}

		ocd := &ocdList.Items[ocdIdx]
		cond := getNotUpgradeableCond(ocd)
		if cond != nil {
			// operator is not upgradeable
			msg := fmt.Sprintf("%s:%s", ocd.GetName(), cond.Message)
			logger.Info("setting operator upgradeable status", "status", cond.Status)
			return r.operatorCondition.Set(context.TODO(), cond.Status,
				conditions.WithReason(cond.Reason), conditions.WithMessage(msg))
		}
	}

	// all operators are upgradeable
	status := metav1.ConditionTrue
	logger.Info("setting operator upgradeable status", "status", status)
	return r.operatorCondition.Set(context.TODO(), status,
		conditions.WithReason("Dependents"), conditions.WithMessage("No dependent reports not upgradeable status"))
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {

	var err error

	r.operatorConditionName, err = util.GetConditionName(mgr.GetClient())
	if err != nil {
		return err
	}
	r.operatorCondition, err = util.NewUpgradeableCondition(mgr.GetClient())
	if err != nil {
		return err
	}

	conditionPredicate := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// not mandatory but these checks wouldn't harm
			if e.ObjectOld == nil || e.ObjectNew == nil {
				return false
			}
			oldObj, _ := e.ObjectOld.(*opv2.OperatorCondition)
			newObj, _ := e.ObjectNew.(*opv2.OperatorCondition)
			if oldObj == nil || newObj == nil {
				return false
			}

			// skip sending a reconcile event if our own condition is updated
			if newObj.GetName() == r.operatorConditionName {
				return false
			}

			// change in admin set conditions for upgradeability
			oldOverride := util.Find(oldObj.Spec.Overrides, func(cond *metav1.Condition) bool {
				return cond.Type == opv2.Upgradeable
			})
			newOverride := util.Find(newObj.Spec.Overrides, func(cond *metav1.Condition) bool {
				return cond.Type == opv2.Upgradeable
			})
			if oldOverride != nil && newOverride == nil {
				// override is removed
				return true
			}
			if newOverride != nil {
				if oldOverride == nil {
					return true
				}
				return oldOverride.Status != newOverride.Status
			}

			// change in operator set conditions for upgradeability
			oldCond := util.Find(oldObj.Status.Conditions, func(cond *metav1.Condition) bool {
				return cond.Type == opv2.Upgradeable
			})
			newCond := util.Find(newObj.Status.Conditions, func(cond *metav1.Condition) bool {
				return cond.Type == opv2.Upgradeable
			})
			if newCond != nil {
				if oldCond == nil {
					return true
				}
				return oldCond.Status != newCond.Status
			}

			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("subscription").
		Watches(
			&opv1a1.Subscription{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Watches(
			&opv2.OperatorCondition{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(conditionPredicate),
		).
		Watches(
			&admrv1.MutatingWebhookConfiguration{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Complete(r)
}

func getNotUpgradeableCond(ocd *opv2.OperatorCondition) *metav1.Condition {
	cond := util.Find(ocd.Spec.Overrides, func(cd *metav1.Condition) bool {
		return cd.Type == opv2.Upgradeable
	})
	if cond != nil {
		if cond.Status != "True" {
			return cond
		}
		// if upgradeable is overridden we should skip checking operator set conditions
		return nil
	}

	return util.Find(ocd.Status.Conditions, func(cd *metav1.Condition) bool {
		return cd.Type == opv2.Upgradeable && cd.Status != "True"
	})
}
