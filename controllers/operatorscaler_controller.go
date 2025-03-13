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
	"go.uber.org/multierr"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

type ResourceMappingRecord struct {
	CrdName    string
	ApiVersion string
	Kind       string
	PkgNames   []string
}

var (
	ResourceMappingList = []ResourceMappingRecord{
		{
			CrdName:    "storageclusters.ocs.openshift.io",
			ApiVersion: "ocs.openshift.io/v1",
			Kind:       "StorageCluster",
			PkgNames:   []string{OcsSubscriptionPackage},
		},
		{
			CrdName:    "cephclusters.ceph.rook.io",
			ApiVersion: "ceph.rook.io/v1",
			Kind:       "CephCluster",
			PkgNames:   []string{RookSubscriptionPackage, CephCSISubscriptionPackage, CSIAddonsSubscriptionPackage},
		},
		{
			CrdName:    "noobaas.noobaa.io",
			ApiVersion: "noobaa.io/v1alpha1",
			Kind:       "NooBaa",
			PkgNames:   []string{NoobaaSubscriptionPackage},
		},
		{
			CrdName:    "prometheuses.monitoring.coreos.com",
			ApiVersion: "monitoring.coreos.com/v1",
			Kind:       "Prometheus",
			PkgNames:   []string{PrometheusSubscriptionPackage},
		},
		{
			CrdName:    "alertmanagers.monitoring.coreos.com",
			ApiVersion: "monitoring.coreos.com/v1",
			Kind:       "Alertmanager",
			PkgNames:   []string{PrometheusSubscriptionPackage},
		},
		{
			CrdName:    "flashsystemclusters.odf.ibm.com",
			ApiVersion: "odf.ibm.com/v1alpha1",
			Kind:       "FlashSystemCluster",
			PkgNames:   []string{IbmSubscriptionPackage},
		},
	}

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

type OperatorScalerReconciler struct {
	client.Client

	Scheme            *runtime.Scheme
	OperatorNamespace string

	ctx               context.Context
	log               logr.Logger
	cache             cache.Cache
	controller        controller.Controller
	kindsBeingWatched map[string]bool
}

//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;list;update

//+kubebuilder:rbac:groups=ocs.openshift.io,resources=storageclusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=ceph.rook.io,resources=cephclusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=noobaa.io,resources=noobaas,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=alertmanagers,verbs=get;list;watch
//+kubebuilder:rbac:groups=odf.ibm.com,resources=flashsystemclusters,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *OperatorScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.ctx = ctx
	r.log = log.FromContext(ctx)

	r.log.Info("starting reconcile")

	if err := r.isOdfDependenciesCsvReady(); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileOperators(); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDynamicWatchers(); err != nil {
		return ctrl.Result{}, err
	}

	r.log.Info("reconcile completed successfully")
	return ctrl.Result{}, nil
}

func (r *OperatorScalerReconciler) isOdfDependenciesCsvReady() error {
	r.log.Info("entering isOdfDependenciesCsvReady")

	odfDepsCsv := &opv1a1.ClusterServiceVersion{}

	err := r.Client.Get(r.ctx, types.NamespacedName{Name: OdfDepsSubscriptionStartingCSV, Namespace: r.OperatorNamespace}, odfDepsCsv)
	if err != nil {
		r.log.Error(err, "failed getting odf-deps csv", "csvName", OdfDepsSubscriptionStartingCSV)
		return err
	}

	if odfDepsCsv.Status.Phase != opv1a1.CSVPhaseSucceeded {
		err = fmt.Errorf("csv %s is not in succeeded state", OdfDepsSubscriptionStartingCSV)
		r.log.Error(err, "waiting for csv to be in succeeded state")
		return err
	}

	r.log.Info("successfully completed isOdfDependenciesCsvReady")
	return nil
}

func (r *OperatorScalerReconciler) reconcileOperators() error {
	r.log.Info("entering reconcileOperators")

	var returnErr error

	for i := range ResourceMappingList {
		resourceMapping := &ResourceMappingList[i]

		crList := &metav1.PartialObjectMetadataList{}
		crList.TypeMeta.APIVersion = resourceMapping.ApiVersion
		crList.TypeMeta.Kind = resourceMapping.Kind

		if err := r.Client.List(r.ctx, crList, client.Limit(1)); err != nil {
			if !meta.IsNoMatchError(err) {
				msg := fmt.Sprintf("failed listing %s", resourceMapping.Kind)
				r.log.Error(err, msg)
				multierr.AppendInto(&returnErr, err)
			}

		} else if len(crList.Items) > 0 {

			for _, pkgName := range resourceMapping.PkgNames {
				key := fmt.Sprintf(CsvLabelKey, pkgName, r.OperatorNamespace)

				csvList := &opv1a1.ClusterServiceVersionList{}
				err = r.Client.List(
					r.ctx, csvList,
					client.InNamespace(r.OperatorNamespace),
					client.MatchingLabels(map[string]string{key: ""}),
				)
				if err != nil {
					r.log.Error(err, "failed listing csvs with label", "label", key)
					multierr.AppendInto(&returnErr, err)
				} else {
					for i := range csvList.Items {
						if err = r.updateCsvDeplymentsReplicas(&csvList.Items[i]); err != nil {
							r.log.Error(err, "failed updating csvs replica")
							multierr.AppendInto(&returnErr, err)
						}
					}
				}
			}
		}
	}

	r.log.Info("successfully completed reconcileOperators")
	return returnErr
}

func (r *OperatorScalerReconciler) updateCsvDeplymentsReplicas(csv *opv1a1.ClusterServiceVersion) error {

	var updateRequired bool
	for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		deploymentSpec := &csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec
		if *deploymentSpec.Replicas < 1 {
			deploymentSpec.Replicas = ptr.To(int32(1))
			updateRequired = true
		}
	}

	if updateRequired {
		if err := r.Client.Update(r.ctx, csv); err != nil {
			r.log.Error(err, "failed updating csv replica", "csvName", csv.Name)
			return err
		}
		r.log.Info("csv updated successfully", "csvName", csv.Name)
	}

	return nil
}

func (r *OperatorScalerReconciler) reconcileDynamicWatchers() error {
	r.log.Info("entering reconcileDynamicWatchers")

	for i := range ResourceMappingList {
		resourceMapping := &ResourceMappingList[i]

		if !r.kindsBeingWatched[resourceMapping.Kind] {

			crd := &extv1.CustomResourceDefinition{}
			crd.Name = resourceMapping.CrdName
			if err := r.Client.Get(r.ctx, client.ObjectKeyFromObject(crd), crd); client.IgnoreNotFound(err) != nil {
				r.log.Error(err, "failed getting crd", "crdName", resourceMapping.CrdName)
				return err
			} else if err == nil {
				r.log.Info("adding dynamic watch", "kind", resourceMapping.Kind)

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
					r.log.Error(err, "failed adding dynamic watch", "kind", resourceMapping.Kind)
					return err
				}

				r.kindsBeingWatched[resourceMapping.Kind] = true
			}
		}
	}

	r.log.Info("successfully completed reconcileDynamicWatchers")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OperatorScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// It will keep track or the crds that we care about
	crdMap := map[string]bool{}
	for i := range ResourceMappingList {
		crdName := ResourceMappingList[i].CrdName
		crdMap[crdName] = true
	}

	controller, err := ctrl.NewControllerManagedBy(mgr).
		Named("operatorScaler").
		Watches(
			&extv1.CustomResourceDefinition{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(obj client.Object) bool {
					return crdMap[obj.GetName()]
				}),
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
