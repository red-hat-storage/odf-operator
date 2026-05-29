/*
Copyright 2026 Data Foundation.

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

package controller

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/red-hat-storage/odf-operator/internal/config"
)

type DeployerReconciler struct {
	Client            client.Client
	Scheme            *runtime.Scheme
	OperatorNamespace string
}

type OlmPkgRecord struct {
	/* example
	   package: ocs-operator
	   version: v5.1.0
	   namespace: openshift-storage
	*/

	Package   string
	Version   string
	Namespace string
}

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

func (r *DeployerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("starting reconcile")

	olmPkgRecords := []*OlmPkgRecord{}
	if err := r.loadOdfPkgsConfigMapData(ctx, logger, &olmPkgRecords); err != nil {
		logger.Error(err, "failed to load odf pkgs configmap")
		return ctrl.Result{}, err
	}

	logger.Info("reconcile completed successfully")
	return ctrl.Result{}, nil
}

func (r *DeployerReconciler) loadOdfPkgsConfigMapData(ctx context.Context, logger logr.Logger, olmPkgRecords *[]*OlmPkgRecord) error {

	configmap, err := config.GetConfigMap(ctx, r.Client, logger, OdfOperatorPkgsConfigMapName, r.OperatorNamespace)
	if err != nil {
		return err
	}

	config.ParsePkgsConfigMapRecords(logger, configmap, r.OperatorNamespace, func(record *config.PkgConfigMapRecord, key, rawValue string) {
		if record.Package == "" || record.Version == "" {
			logger.Info("skipping the record from the configmap", "key", key, "value", rawValue)
			return
		}

		*olmPkgRecords = append(*olmPkgRecords, &OlmPkgRecord{
			Package:   record.Package,
			Version:   record.Version,
			Namespace: record.Namespace,
		})
	})

	logger.Info("olm records", "records", olmPkgRecords)

	return nil
}

func (r *DeployerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	odfPkgConfigmapPredicate := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetName() == OdfOperatorPkgsConfigMapName && obj.GetNamespace() == r.OperatorNamespace
	})

	return ctrl.NewControllerManagedBy(mgr).
		Named("deployer").
		Watches(
			&corev1.ConfigMap{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(
				odfPkgConfigmapPredicate,
				predicate.GenerationChangedPredicate{},
			),
		).
		Complete(r)
}
