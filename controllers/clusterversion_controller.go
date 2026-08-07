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
	ocstlsv1 "github.com/red-hat-storage/ocs-tls-profiles/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/red-hat-storage/odf-operator/console"
	"github.com/red-hat-storage/odf-operator/pkg/util"
)

const rotatedVersionAnnotationKey = "odf.openshift.io/rotated-version"

// ClusterVersionReconciler reconciles a ClusterVersion object
type ClusterVersionReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	ConsolePort     int32
	cache           cache.Cache
	controller      controller.Controller
	tlsWatchStarted bool
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
// OCP certification requires these permissions even though bundle manifests install the policies.
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=create;update;delete
//+kubebuilder:rbac:groups=ocs.openshift.io,resources=tlsprofiles,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ClusterVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	r.ensureTLSProfileWatch(ctx)
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

	controller, err := ctrl.NewControllerManagedBy(mgr).
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
		Watches(
			&extv1.CustomResourceDefinition{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(obj client.Object) bool {
					return obj.GetName() == "tlsprofiles.ocs.openshift.io"
				}),
			),
		).
		Build(r)

	r.controller = controller
	r.cache = mgr.GetCache()

	return err
}

func (r *ClusterVersionReconciler) ensureTLSProfileWatch(ctx context.Context) {
	if r.tlsWatchStarted {
		return
	}
	logger := log.FromContext(ctx)

	if err := ocstlsv1.AddToScheme(r.Scheme); err != nil {
		return
	}

	crd := &extv1.CustomResourceDefinition{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: "tlsprofiles.ocs.openshift.io"}, crd); err != nil {
		logger.Info("TLSProfile CRD not available yet, will retry on next reconcile")
		return
	}

	if err := r.controller.Watch(
		source.Kind(
			r.cache,
			&ocstlsv1.TLSProfile{},
			&handler.TypedEnqueueRequestForObject[*ocstlsv1.TLSProfile]{},
			predicate.And(
				predicate.NewTypedPredicateFuncs(func(obj *ocstlsv1.TLSProfile) bool {
					return obj.GetName() == TLSProfileName && obj.GetNamespace() == OperatorNamespace
				}),
				predicate.TypedGenerationChangedPredicate[*ocstlsv1.TLSProfile]{},
			),
		),
	); err != nil {
		logger.Error(err, "Failed to add TLSProfile watch")
		return
	}
	r.tlsWatchStarted = true
	logger.Info("Dynamic watch added for TLSProfile")
}

func (r *ClusterVersionReconciler) ensureConsolePlugin(ctx context.Context, clusterVersion string) error {
	logger := log.FromContext(ctx)
	// The base path to where the request are sent
	basePath := console.GetBasePath(clusterVersion)

	var ossl *ocstlsv1.OpenSSLConfig
	if r.tlsWatchStarted {
		tlsProfile := &ocstlsv1.TLSProfile{}
		tlsProfile.Name = TLSProfileName
		tlsProfile.Namespace = OperatorNamespace
		if err := r.Client.Get(ctx, client.ObjectKeyFromObject(tlsProfile), tlsProfile); client.IgnoreNotFound(err) != nil {
			return err
		}
		if cfg, found := ocstlsv1.GetConfigForServer(tlsProfile, "odf.openshift.io", "console"); found {
			if err := ocstlsv1.ValidateTLSConfig(cfg); err != nil {
				logger.Error(err, "Invalid TLSProfile config, using nginx defaults")
			} else {
				ossl = ocstlsv1.OpenSSLConfigFrom(ocstlsv1.GetGoTLSConfig(cfg))
			}
		}
	}

	nginxConf := console.GenerateNginxConf(ossl)

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

	// Create/Update ODF console ConfigMap (nginx configuration).
	// The ConfigMap is mounted as a directory volume in the console pod.
	// A background watcher in the container detects changes and reloads nginx.
	odfConsoleConfigMap := console.GetNginxConfConfigMap(OperatorNamespace)
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, odfConsoleConfigMap, func() error {
		if odfConsoleConfigMap.Data == nil {
			odfConsoleConfigMap.Data = make(map[string]string)
		}
		odfConsoleConfigMap.Data["nginx.conf"] = nginxConf
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
		// MergeConsolePluginProxy merges the desired proxy entries into the existing.
		// To remove stale proxy entries, pass the aliases of the stale entries to the function (currently nil)
		odfConsolePlugin.Spec.Proxy = console.MergeConsolePluginProxy(odfConsolePlugin.Spec.Proxy, OperatorNamespace, nil)
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
	odfMajorMinorVersion := fmt.Sprintf("%d.%d", odfCsv.Spec.Version.Major, odfCsv.Spec.Version.Minor)
	if _, err = controllerutil.CreateOrUpdate(ctx, r.Client, uxBackendServerSecret, func() error {
		uxBackendServerSecret.SetOwnerReferences(nil)
		currentVersion := uxBackendServerSecret.Annotations[rotatedVersionAnnotationKey]
		if currentVersion != odfMajorMinorVersion {
			secret, err := generateSessionSecret()
			if err != nil {
				return fmt.Errorf("failed to generate session secret: %w", err)
			}
			uxBackendServerSecret.StringData = map[string]string{
				"session_secret": secret,
			}
			if uxBackendServerSecret.Annotations == nil {
				uxBackendServerSecret.Annotations = make(map[string]string)
			}
			uxBackendServerSecret.Annotations[rotatedVersionAnnotationKey] = odfMajorMinorVersion
		}
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

	// Get tolerations from ODF subscription for deployment
	var tolerations []corev1.Toleration
	if odfSub, err := GetOdfSubscription(ctx, r.Client); err != nil {
		return fmt.Errorf("failed to get ODF subscription: %w", err)
	} else if odfSub.Spec.Config != nil {
		tolerations = odfSub.Spec.Config.Tolerations
	}

	// Create/Update UX backend server deployment
	logger.Info("Ensuring UX backend server deployment")
	uxBackendServerDeployment := getUXBackendServerDeployment(tolerations)
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
