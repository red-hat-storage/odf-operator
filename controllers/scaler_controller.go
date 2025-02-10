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

	"github.com/go-logr/logr"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ScalerReconciler struct {
	ctx               context.Context
	logger            logr.Logger
	Client            client.Client
	Scheme            *runtime.Scheme
	OperatorNamespace string
	controller        controller.Controller
	mgr               ctrl.Manager
}

var (
	kindStorageCluster = &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocs.openshift.io/v1",
			Kind:       "StorageCluster",
		},
	}
	kindCephCluster = &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ceph.rook.io/v1",
			Kind:       "CephCluster",
		},
	}
	kindFlashSystemCluster = &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "odf.ibm.com/v1alpha1",
			Kind:       "FlashSystemCluster",
		},
	}
)

var (
	kinds = []metav1.PartialObjectMetadata{
		*kindStorageCluster,
		*kindCephCluster,
		*kindFlashSystemCluster,
	}
)

//+kubebuilder:rbac:groups=ocs.openshift.io,resources=storageclusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=ceph.rook.io,resources=cephclusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=odf.ibm.com,resources=flashsystemclusters,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.ctx = ctx
	r.logger = log.FromContext(ctx)

	r.logger.Info("starting reconcile")

	if req.Name == "storageclusters.ocs.openshift.io" {
		r.logger.Info("adding watch for storageclusters")
		err := r.addWatch(kindStorageCluster)
		if err != nil {
			r.logger.Error(err, "failed to add watch")
			return ctrl.Result{}, err
		}
	}
	if req.Name == "cephclusters.ceph.rook.io" {
		r.logger.Info("adding watch for cephclusters")
		err := r.addWatch(kindCephCluster)
		if err != nil {
			r.logger.Error(err, "failed to add watch")
			return ctrl.Result{}, err
		}
	}
	if req.Name == "flashsystemclusters.odf.ibm.com" {
		r.logger.Info("adding watch for flashsystemclusters")
		err := r.addWatch(kindFlashSystemCluster)
		if err != nil {
			r.logger.Error(err, "failed to add watch")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ScalerReconciler) addWatch(kind *metav1.PartialObjectMetadata) error {
	return r.controller.Watch(
		source.Kind(
			r.mgr.GetCache(),
			client.Object(kind),
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}}}
				},
			),
		),
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	controller, err := ctrl.NewControllerManagedBy(mgr).
		Named("scaler").
		Watches(
			&extv1.CustomResourceDefinition{},
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					name := obj.GetName()
					if name == "storageclusters.ocs.openshift.io" ||
						name == "cephclusters.ceph.rook.io" ||
						name == "flashsystemclusters.odf.ibm.com" {
						return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: name}}}
					}
					return []reconcile.Request{}
				},
			)).
		Build(r)

	r.controller = controller
	r.mgr = mgr

	return err
}
