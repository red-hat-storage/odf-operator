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

	configv1 "github.com/openshift/api/config/v1"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8sUtil "k8s.io/client-go/tools/record/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/red-hat-storage/odf-operator/console"
	"github.com/red-hat-storage/odf-operator/pkg/util"
)

// ClusterVersionReconciler reconciles a ClusterVersion object
type ClusterVersionReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	ConsolePort int
}

//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;create;update;delete
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleclidownloads,verbs=get;create;update;delete

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ClusterVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	instance := configv1.ClusterVersion{}
	if err := r.Client.Get(context.TODO(), req.NamespacedName, &instance); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureConsolePlugin(instance.Status.Desired.Version); err != nil {
		logger.Error(err, "Could not ensure compatibility for ODF consolePlugin")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterVersionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.Add(manager.RunnableFunc(func(context.Context) error {
		clusterVersion, err := util.DetermineOpenShiftVersion(r.Client)
		if err != nil {
			return err
		}

		return r.ensureConsolePlugin(clusterVersion)
	}))
	if err != nil {
		return err
	}

	enqueueClusterVersionRequest := handler.EnqueueRequestsFromMapFunc(
		func(context context.Context, obj client.Object) []reconcile.Request {
			logger := log.FromContext(context)

			instance, _ := obj.(*operatorv1alpha1.Subscription)
			logger.Info("TEST TEST TEST TEST: Inside enqueueClusterVersionRequest")
			logger.Info(fmt.Sprintf("TEST TEST TEST TEST: obj Kind '%s'", instance.Kind))
			logger.Info(fmt.Sprintf("TEST TEST TEST TEST: obj deletionTimeStamp '%s'", instance.ObjectMeta.DeletionTimestamp))

			// Get the ClusterVersion objects
			clusterVersionList := &configv1.ClusterVersionList{}
			err := r.Client.List(context, clusterVersionList)
			if err != nil {
				logger.Error(err, "Unable to list ClusterVersion objects")
				return []reconcile.Request{}
			}

			// Return name and namespace of the ClusterVersion object
			request := []reconcile.Request{}
			for _, cv := range clusterVersionList.Items {
				request = append(request, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: cv.Namespace,
						Name:      cv.Name,
					},
				})
			}

			return request
		},
	)

	subscriptionPredicate := predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		UpdateFunc:  func(e event.UpdateEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
		DeleteFunc: func(e event.DeleteEvent) bool {
			instance, ok := e.Object.(*operatorv1alpha1.Subscription)
			if !ok {
				return false
			}

			// ignore if not a odf-operator subscription
			if instance.Spec.Package != "odf-operator" || instance.Namespace != OperatorNamespace {
				return false
			}

			// for deleting cluster-scoped resources during uninstall
			return true
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1.ClusterVersion{}).
		Watches(
			&operatorv1alpha1.Subscription{},
			enqueueClusterVersionRequest,
			builder.WithPredicates(subscriptionPredicate),
		).
		Complete(r)
}

func (r *ClusterVersionReconciler) ensureConsolePlugin(clusterVersion string) error {
	logger := log.FromContext(context.TODO())
	// The base path to where the request are sent
	basePath := console.GetBasePath(clusterVersion)
	nginxConf := console.NginxConf

	// Customer portal link (CLI Tool download)
	portalLink := console.CUSTOMER_PORTAL_LINK

	logger.Info("TEST TEST TEST TEST: Inside ensureConsolePlugin")

	// Get ODF console Deployment
	odfConsoleDeployment := console.GetDeployment(OperatorNamespace)
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      odfConsoleDeployment.Name,
		Namespace: odfConsoleDeployment.Namespace,
	}, odfConsoleDeployment)
	if err != nil {
		return err
	}

	logger.Info("TEST TEST TEST TEST: Fetched the Deployment")

	sub, err := GetOdfSubscription(r.Client)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "failed to get ODF Subscription")
		logger.Info(fmt.Sprintf("TEST TEST TEST TEST: Subscription err reason '%s'", errors.ReasonForError(err)))
		logger.Info(fmt.Sprintf("TEST TEST TEST TEST: Subscription err code is 404 ?? '%t'", k8sUtil.IsKeyNotFoundError(err)))
		return err
	}

	logger.Info("TEST TEST TEST TEST: Fetched the Subscription")
	logger.Info(fmt.Sprintf("TEST TEST TEST TEST: Subscription name '%s'", sub.Name))
	logger.Info(fmt.Sprintf("TEST TEST TEST TEST: Subscription namespace '%s'", sub.Namespace))
	logger.Info(fmt.Sprintf("TEST TEST TEST TEST: Subscription deletionTimeStamp '%s'", sub.ObjectMeta.DeletionTimestamp))

	// Deleting cluster-scoped resources which cannot be garbage collected,
	// when either Subscription is marked for deletion or already deleted during this block's execution
	if errors.IsNotFound(err) || k8sUtil.IsKeyNotFoundError(err) || !sub.GetDeletionTimestamp().IsZero() {

		logger.Info("TEST TEST TEST TEST: Inside deletion condition")
		logger.Info(fmt.Sprintf("TEST TEST TEST TEST: errors.IsNotFound(err) ?? '%t'", errors.IsNotFound(err)))
		logger.Info(fmt.Sprintf("TEST TEST TEST TEST: k8sUtil.IsKeyNotFoundError(err) ?? '%t'", k8sUtil.IsKeyNotFoundError(err)))

		odfConsolePlugin := console.GetConsolePluginCR(r.ConsolePort, OperatorNamespace)
		err := r.Client.Delete(context.TODO(), odfConsolePlugin)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "Failed to delete ConsolePlugin", "Name", odfConsolePlugin.ObjectMeta.Name)
			return err
		}
		logger.Info("ConsolePlugin deleted successfully", "Name", odfConsolePlugin.ObjectMeta.Name)

		consoleCLIDownload := console.GetConsoleCLIDownloadCR()
		err = r.Client.Delete(context.TODO(), consoleCLIDownload)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "Failed to delete ConsoleCLIDownload", "Name", consoleCLIDownload.ObjectMeta.Name)
			return err
		}
		logger.Info("ConsoleCLIDownload deleted successfully", "Name", consoleCLIDownload.ObjectMeta.Name)

		// return early if deletion successful
		return nil
	}

	// Create/Update ODF console ConfigMap (nginx configuration)
	odfConsoleConfigMap := console.GetNginxConfConfigMap(OperatorNamespace)
	_, err = controllerutil.CreateOrUpdate(context.TODO(), r.Client, odfConsoleConfigMap, func() error {
		if odfConsoleConfigMapData := odfConsoleConfigMap.Data["nginx.conf"]; odfConsoleConfigMapData != nginxConf {
			logger.Info(fmt.Sprintf("Set the ConfigMap odf-console-nginx-conf data as '%s'", nginxConf))
			odfConsoleConfigMap.Data["nginx.conf"] = nginxConf
		}
		return controllerutil.SetControllerReference(odfConsoleDeployment, odfConsoleConfigMap, r.Scheme)
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	// Create/Update ODF console Service
	odfConsoleService := console.GetService(r.ConsolePort, OperatorNamespace)
	_, err = controllerutil.CreateOrUpdate(context.TODO(), r.Client, odfConsoleService, func() error {
		return controllerutil.SetControllerReference(odfConsoleDeployment, odfConsoleService, r.Scheme)
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	// Create/Update ODF console ConsolePlugin
	odfConsolePlugin := console.GetConsolePluginCR(r.ConsolePort, OperatorNamespace)
	_, err = controllerutil.CreateOrUpdate(context.TODO(), r.Client, odfConsolePlugin, func() error {
		if currentBasePath := odfConsolePlugin.Spec.Service.BasePath; currentBasePath != basePath {
			logger.Info(fmt.Sprintf("Set the BasePath for odf-console plugin as '%s'", basePath))
			odfConsolePlugin.Spec.Service.BasePath = basePath
		}
		if odfConsolePlugin.Spec.Proxy == nil {
			odfConsolePlugin.Spec.Proxy = console.GetConsolePluginProxy(OperatorNamespace)
		}
		return nil
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	// Create/Update ConsoleCLIDownload (CLI Tool download)
	consoleCLIDownload := console.GetConsoleCLIDownloadCR()
	_, err = controllerutil.CreateOrUpdate(context.TODO(), r.Client, consoleCLIDownload, func() error {
		if currentPortalLink := consoleCLIDownload.Spec.Links[0].Href; currentPortalLink != portalLink {
			logger.Info(fmt.Sprintf("Set the customer portal link for CLI Tool '%s'", portalLink))
			consoleCLIDownload.Spec.Links[0].Href = portalLink
		}
		if len(consoleCLIDownload.Spec.Links) != 1 {
			consoleCLIDownload.Spec.Links = console.GetConsoleCLIDownloadLinks()
		}
		return nil
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	logger.Info("TEST TEST TEST TEST: returning from ensureConsolePlugin")

	return nil
}
