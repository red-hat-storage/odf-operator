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

	"github.com/go-logr/logr"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/red-hat-storage/odf-operator/metrics"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// csvLabelKey is a label key to identify CSV with pkg name and namespace
	CsvLabelKey = "operators.coreos.com/%s.%s"
)

var (
	createOnlyPredicate = predicate.Funcs{
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
)

type KindPackagesRecord struct {
	/* examples
	   ApiVersion: "ceph.rook.io/v1",
	   Kind:       "CephCluster",
	   PkgNames:   []string{rook-operator, cephcsi-operator, csi-addons, ocs-client-operator},
	*/

	ApiVersion string
	Kind       string
	PkgNames   []string
}

type OperatorScalerReconciler struct {
	client.Client

	Scheme            *runtime.Scheme
	OperatorNamespace string

	cache             cache.Cache
	controller        controller.Controller
	kindsBeingWatched map[string]bool
}

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;list;update

//+kubebuilder:rbac:groups=ocs.openshift.io,resources=storageclusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=ceph.rook.io,resources=cephclusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=noobaa.io,resources=noobaas,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=alertmanagers,verbs=get;list;watch
//+kubebuilder:rbac:groups=odf.ibm.com,resources=flashsystemclusters,verbs=get;list;watch

func (r *OperatorScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("starting reconcile")

	var kindMapping = map[string]*KindPackagesRecord{}
	var odfDepsCsvName = ""
	if err := r.loadOdfConfigMapData(ctx, logger, kindMapping, &odfDepsCsvName); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.isOdfDependenciesCsvReady(ctx, logger, odfDepsCsvName); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileOperators(ctx, logger, kindMapping); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDynamicWatchers(ctx, logger, kindMapping); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMetrics(ctx, logger); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconcile completed successfully")
	return ctrl.Result{}, nil
}

func (r *OperatorScalerReconciler) loadOdfConfigMapData(ctx context.Context, logger logr.Logger, kindMapping map[string]*KindPackagesRecord, odfDepsCsvName *string) error {
	logger.Info("entering loadOdfConfigMapData")

	configmap, err := GetOdfConfigMap(ctx, r.Client, logger)
	if err != nil {
		return err
	}

	var combinedErr error

	ParseOdfConfigMapRecords(logger, configmap, func(record *OdfOperatorConfigMapRecord, key, rawValue string) {
		if record.Pkg == OdfDepsSubscriptionPackage {
			*odfDepsCsvName = record.Csv
		}

		if record.Pkg == "" || record.ScaleUpOnInstanceOf == nil {
			logger.Info("skipping the record from the configmap", "key", key, "value", rawValue)
			return
		}
		for _, crdName := range record.ScaleUpOnInstanceOf {

			rec, ok := kindMapping[crdName]
			if !ok {
				rec = &KindPackagesRecord{}
			}
			rec.PkgNames = append(rec.PkgNames, record.Pkg)

			// populate the apiVersion and kind
			crd := &extv1.CustomResourceDefinition{}
			crd.Name = crdName
			if err := r.Client.Get(ctx, client.ObjectKeyFromObject(crd), crd); errors.IsNotFound(err) {
				logger.Info("skipping crd not found", "crdName", crdName)
				continue
			} else if err != nil {
				logger.Error(err, "failed getting crd", "crdName", crdName)
				multierr.AppendInto(&combinedErr, err)
				continue
			}

			rec.ApiVersion = crd.Spec.Group + "/" + crd.Spec.Versions[0].Name
			rec.Kind = crd.Spec.Names.Kind
			kindMapping[crdName] = rec
		}
	})

	logger.Info("operator scaler records", "records", kindMapping)

	if combinedErr == nil {
		logger.Info("successfully completed loadOdfConfigMapData")
	}

	return combinedErr
}

func (r *OperatorScalerReconciler) isOdfDependenciesCsvReady(ctx context.Context, logger logr.Logger, odfDepsCsvName string) error {
	logger.Info("entering isOdfDependenciesCsvReady")

	odfDepsCsv := &opv1a1.ClusterServiceVersion{}

	err := r.Client.Get(ctx, types.NamespacedName{Name: odfDepsCsvName, Namespace: r.OperatorNamespace}, odfDepsCsv)
	if err != nil {
		logger.Error(err, "failed getting odf-deps csv", "csvName", odfDepsCsvName)
		return err
	}

	if odfDepsCsv.Status.Phase != opv1a1.CSVPhaseSucceeded {
		err = fmt.Errorf("csv %s is not in succeeded state", odfDepsCsvName)
		logger.Error(err, "waiting for csv to be in succeeded state")
		return err
	}

	logger.Info("successfully completed isOdfDependenciesCsvReady")
	return nil
}

func (r *OperatorScalerReconciler) reconcileMetrics(ctx context.Context, logger logr.Logger) error {
	logger.Info("entering reconcileMetrics")

	// list the crds with label and update the metrics
	crdList := &extv1.CustomResourceDefinitionList{}
	labelOptions := client.MatchingLabels{"odf.openshift.io/is-storage-system": "true"}
	if err := r.Client.List(ctx, crdList, labelOptions); err != nil {
		logger.Error(err, "failed to list CRDs with label", "label", labelOptions)
		return err
	}

	var combinedErr error
	for i := range crdList.Items {
		crd := &crdList.Items[i]

		crList := &metav1.PartialObjectMetadataList{}
		crList.APIVersion = crd.Spec.Group + "/" + crd.Spec.Versions[0].Name
		crList.Kind = crd.Spec.Names.Kind

		if err := r.Client.List(ctx, crList); err != nil {
			msg := fmt.Sprintf("failed listing %s", crList.Kind)
			logger.Error(err, msg)
			multierr.AppendInto(&combinedErr, err)
		} else {
			for j := range crList.Items {
				crItem := &crList.Items[j]

				metrics.ReportODFSystemMapMetrics(
					crItem.Name,
					crItem.Name,
					crItem.Namespace,
					crList.Kind,
					crList.APIVersion,
				)
			}
		}
	}
	if combinedErr == nil {
		logger.Info("successfully completed reconcileMetrics")
	}

	return combinedErr
}

func (r *OperatorScalerReconciler) reconcileOperators(ctx context.Context, logger logr.Logger, kindMapping map[string]*KindPackagesRecord) error {
	logger.Info("entering reconcileOperators")

	var returnErr error

	for _, resourceMapping := range kindMapping {

		crList := &metav1.PartialObjectMetadataList{}
		crList.TypeMeta.APIVersion = resourceMapping.ApiVersion
		crList.TypeMeta.Kind = resourceMapping.Kind

		if err := r.Client.List(ctx, crList, client.Limit(1)); err != nil {
			if !meta.IsNoMatchError(err) {
				msg := fmt.Sprintf("failed listing %s", resourceMapping.Kind)
				logger.Error(err, msg)
				multierr.AppendInto(&returnErr, err)
			}

		} else if len(crList.Items) > 0 {

			for _, pkgName := range resourceMapping.PkgNames {
				key := fmt.Sprintf(CsvLabelKey, pkgName, r.OperatorNamespace)

				csvList := &opv1a1.ClusterServiceVersionList{}
				err = r.Client.List(
					ctx, csvList,
					client.InNamespace(r.OperatorNamespace),
					client.MatchingLabels(map[string]string{key: ""}),
				)
				if err != nil {
					logger.Error(err, "failed listing csvs with label", "label", key)
					multierr.AppendInto(&returnErr, err)
				} else {
					for j := range csvList.Items {
						if err = r.updateCsvDeplymentsReplicas(ctx, logger, &csvList.Items[j]); err != nil {
							logger.Error(err, "failed updating csvs replica")
							multierr.AppendInto(&returnErr, err)
						}
					}
				}
			}
		}
	}

	if returnErr == nil {
		logger.Info("successfully completed reconcileOperators")
	}

	return returnErr
}

func (r *OperatorScalerReconciler) updateCsvDeplymentsReplicas(ctx context.Context, logger logr.Logger, csv *opv1a1.ClusterServiceVersion) error {

	var updateRequired bool
	for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		deploymentSpec := &csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec
		if deploymentSpec.Replicas == nil || *deploymentSpec.Replicas < 1 {
			deploymentSpec.Replicas = ptr.To(int32(1))
			updateRequired = true
		}
	}

	if updateRequired {
		if err := r.Client.Update(ctx, csv); err != nil {
			logger.Error(err, "failed updating csv replica", "csvName", csv.Name)
			return err
		}
		logger.Info("csv updated successfully", "csvName", csv.Name)
	}

	return nil
}

func (r *OperatorScalerReconciler) reconcileDynamicWatchers(_ context.Context, logger logr.Logger, kindMapping map[string]*KindPackagesRecord) error {
	logger.Info("entering reconcileDynamicWatchers")

	for _, resourceMapping := range kindMapping {

		if !r.kindsBeingWatched[resourceMapping.Kind] {

			logger.Info("adding dynamic watch", "kind", resourceMapping.Kind)

			err := r.controller.Watch(
				source.Kind(
					r.cache,
					client.Object(
						&metav1.PartialObjectMetadata{
							TypeMeta: metav1.TypeMeta{
								APIVersion: resourceMapping.ApiVersion,
								Kind:       resourceMapping.Kind,
							},
						},
					),
					&handler.EnqueueRequestForObject{},
					// Trigger the reconcile for creation events of the object.
					// This ensures the replicas in the CSV are scaled up based on the presence of Custom Resource (CR).
					createOnlyPredicate,
				),
			)
			if err != nil {
				logger.Error(err, "failed adding dynamic watch", "kind", resourceMapping.Kind)
				return err
			}

			r.kindsBeingWatched[resourceMapping.Kind] = true
		}
	}

	logger.Info("successfully completed reconcileDynamicWatchers")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OperatorScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	controller, err := ctrl.NewControllerManagedBy(mgr).
		Named("operatorScaler").
		Watches(
			&corev1.ConfigMap{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(obj client.Object) bool {
					return obj.GetName() == odfOperatorConfigMapName && obj.GetNamespace() == r.OperatorNamespace
				}),
				predicate.GenerationChangedPredicate{},
			),
		).
		Watches(
			&extv1.CustomResourceDefinition{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(
				//TODO: filter the crds we care about
				// Trigger a reconcile only during the creation of a specific CRD to ensure it runs exactly once for that CRD.
				// This is required to dynamically add a watch for the corresponding Custom Resource (CR) based on the CRD name.
				// The reconcile will be triggered with the CRD name as `req.Name`, and the reconciler will set up a watch for the CR using the CRD name.
				createOnlyPredicate,
			),
		).
		Build(r)

	r.kindsBeingWatched = map[string]bool{}
	r.controller = controller
	r.cache = mgr.GetCache()

	return err
}
