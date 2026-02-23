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
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/red-hat-storage/odf-operator/console"
	"github.com/red-hat-storage/odf-operator/pkg/util"
)

// ClusterVersionReconciler reconciles a ClusterVersion object
type ClusterVersionReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	ConsolePort int32
}

//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;create;update
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleclidownloads,verbs=get;create;update
//+kubebuilder:rbac:groups=console.openshift.io,resources=consolequickstarts,verbs=get;list;create;update;delete

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ClusterVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	ocpVersion, err := util.GetOpenShiftVersion(ctx, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureConsolePlugin(ctx, ocpVersion); err != nil {
		logger.Error(err, "Could not ensure compatibility for ODF consolePlugin")
		return ctrl.Result{}, err
	}

	if err := r.ensureUXBackendServer(ctx); err != nil {
		logger.Error(err, "Could not ensure UX backend server")
		return ctrl.Result{}, err
	}

	if err := ensureQuickStarts(ctx, r.Client, logger); err != nil {
		logger.Error(err, "Could not ensure QuickStarts")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterVersionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		clusterVersion, err := util.GetOpenShiftVersion(ctx, r.Client)
		if err != nil {
			return err
		}

		return r.ensureConsolePlugin(ctx, clusterVersion)
	}))
	if err != nil {
		return err
	}

	uxBackendResourcePredicate := func(name string) predicate.Predicate {
		return predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetName() == name && obj.GetNamespace() == OperatorNamespace
		})
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1.ClusterVersion{}).
		Watches(
			&appsv1.Deployment{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(uxBackendResourcePredicate("ux-backend-server")),
		).
		Watches(
			&corev1.Secret{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(uxBackendResourcePredicate("ux-backend-proxy")),
		).
		Watches(
			&corev1.Service{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(uxBackendResourcePredicate("ux-backend-proxy")),
		).
		Complete(r)
}

func (r *ClusterVersionReconciler) ensureConsolePlugin(ctx context.Context, clusterVersion string) error {
	logger := log.FromContext(ctx)
	// The base path to where the request are sent
	basePath := console.GetBasePath(clusterVersion)
	nginxConf := console.NginxConf

	// Customer portal link (CLI Tool download)
	portalLink := console.CUSTOMER_PORTAL_LINK

	// Get ODF console Deployment
	odfConsoleDeployment := console.GetDeployment(OperatorNamespace)
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      odfConsoleDeployment.Name,
		Namespace: odfConsoleDeployment.Namespace,
	}, odfConsoleDeployment)
	if err != nil {
		return err
	}

	// Create/Update ODF console ConfigMap (nginx configuration)
	odfConsoleConfigMap := console.GetNginxConfConfigMap(OperatorNamespace)
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, odfConsoleConfigMap, func() error {
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
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, odfConsoleService, func() error {
		return controllerutil.SetControllerReference(odfConsoleDeployment, odfConsoleService, r.Scheme)
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	// Create/Update ODF console ConsolePlugin
	odfConsolePlugin := console.GetConsolePluginCR(r.ConsolePort, OperatorNamespace)
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, odfConsolePlugin, func() error {
		if odfConsolePlugin.Spec.Backend.Service != nil {
			if currentBasePath := odfConsolePlugin.Spec.Backend.Service.BasePath; currentBasePath != basePath {
				logger.Info(fmt.Sprintf("Set the BasePath for odf-console plugin as '%s'", basePath))
				odfConsolePlugin.Spec.Backend.Service.BasePath = basePath
			}
		}
		odfConsolePlugin.Spec.Proxy = console.GetConsolePluginProxy(OperatorNamespace)
		return nil
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	// Create/Update ConsoleCLIDownload (CLI Tool download)
	consoleCLIDownload := console.GetConsoleCLIDownloadCR()
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, consoleCLIDownload, func() error {
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

	return nil
}

func (r *ClusterVersionReconciler) ensureUXBackendServer(ctx context.Context) error {
	logger := log.FromContext(ctx)
	odfCsvName, err := util.GetConditionName(r.Client)
	if err != nil {
		return fmt.Errorf("failed to get ODF CSV name: %w", err)
	}

	odfCsv := &opv1a1.ClusterServiceVersion{}
	odfCsv.Name = odfCsvName
	odfCsv.Namespace = OperatorNamespace
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(odfCsv), odfCsv); err != nil {
		return fmt.Errorf("failed to get ODF CSV %s/%s: %w", odfCsv.Namespace, odfCsv.Name, err)
	}

	// TODO: remove the following check in future version
	// the following is to check if ocs-operator and odf-operator csvs are at same version
	ocsCsvName := strings.Replace(odfCsvName, "odf", "ocs", 1)
	ocsCSV := &opv1a1.ClusterServiceVersion{}
	ocsCSV.Name = ocsCsvName
	ocsCSV.Namespace = OperatorNamespace
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(ocsCSV), ocsCSV); err != nil {
		// The OCS operator CSV must match the ODF operator CSV version,
		// this is because the ux-backend-server deployment is moved from ocs-operator to odf-operator
		// during upgrade there may be a collision of ownership between the two operators
		logger.Error(err, "Skipping UX backend server setup")
		return fmt.Errorf("OCS operator CSV must match the ODF operator CSV version: %w", err)
	}

	logger.Info("Ensuring UX backend server secret")
	uxBackendServerSecret := getUXBackendServerSecret()
	secretData := uxBackendServerSecret.StringData
	if _, err = controllerutil.CreateOrUpdate(ctx, r.Client, uxBackendServerSecret, func() error {
		uxBackendServerSecret.SetOwnerReferences(nil)
		uxBackendServerSecret.StringData = secretData
		return controllerutil.SetControllerReference(odfCsv, uxBackendServerSecret, r.Scheme)
	}); err != nil {
		return fmt.Errorf("failed to create or update UX backend server secret: %w", err)
	}

	logger.Info("Ensuring UX backend server service")
	uxBackendServerService := getUXBackendServerService()
	desiredService := uxBackendServerService.Spec.DeepCopy()
	annotations := uxBackendServerService.Annotations
	if _, err = controllerutil.CreateOrUpdate(ctx, r.Client, uxBackendServerService, func() error {
		uxBackendServerService.SetOwnerReferences(nil)
		uxBackendServerService.Spec = *desiredService
		uxBackendServerService.Annotations = annotations
		return controllerutil.SetControllerReference(odfCsv, uxBackendServerService, r.Scheme)
	}); err != nil {
		return fmt.Errorf("failed to create or update UX backend server service: %w", err)
	}

	// Create/Update UX backend server deployment
	logger.Info("Ensuring UX backend server deployment")
	uxBackendServerDeployment := getUXBackendServerDeployment()
	desiredSpec := uxBackendServerDeployment.Spec.DeepCopy()
	if _, err = controllerutil.CreateOrUpdate(ctx, r.Client, uxBackendServerDeployment, func() error {
		uxBackendServerDeployment.SetOwnerReferences(nil)
		uxBackendServerDeployment.Spec = *desiredSpec
		return controllerutil.SetControllerReference(odfCsv, uxBackendServerDeployment, r.Scheme)
	}); err != nil {
		return fmt.Errorf("failed to create or update UX backend server deployment: %w", err)
	}

	return nil
}
