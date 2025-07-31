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

package main

import (
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	opv2 "github.com/operator-framework/api/pkg/operators/v2"
	admrv1 "k8s.io/api/admissionregistration/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	odfv1a1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	"github.com/red-hat-storage/odf-operator/controllers"
	"github.com/red-hat-storage/odf-operator/pkg/util"
	"github.com/red-hat-storage/odf-operator/webhook"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(k8sscheme.AddToScheme(scheme))
	utilruntime.Must(odfv1a1.AddToScheme(scheme))
	utilruntime.Must(opv1a1.AddToScheme(scheme))
	utilruntime.Must(opv2.AddToScheme(scheme))
	utilruntime.Must(consolev1.AddToScheme(scheme))
	utilruntime.Must(admrv1.AddToScheme(scheme))
	utilruntime.Must(extv1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var odfConsolePort int
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8085", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8082", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&odfConsolePort, "odf-console-port", 9001, "The port where the ODF console server will be serving it's payload")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	operatorNamespace, err := util.GetOperatorNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get operator namespace")
		os.Exit(1)
	}

	cacheOptions, err := getCacheOptions(operatorNamespace)
	if err != nil {
		setupLog.Error(err, "unable to get cache options")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metrics.Options{
			BindAddress:    metricsAddr,
			SecureServing:  true,
			FilterProvider: filters.WithAuthenticationAndAuthorization,
		},
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "4fd470de.openshift.io",
		LeaderElectionNamespace: operatorNamespace,
		Cache:                   cacheOptions,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ClusterVersionReconciler{
		Client:      mgr.GetClient(),
		Scheme:      mgr.GetScheme(),
		ConsolePort: int32(odfConsolePort), //nolint:gosec
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterVersion")
		os.Exit(1)
	}

	if err = (&controllers.SubscriptionReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		OperatorNamespace: operatorNamespace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Subscription")
		os.Exit(1)
	}

	if err = (&controllers.OperatorScalerReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		OperatorNamespace: operatorNamespace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OperatorScaler")
		os.Exit(1)
	}

	if err = (&controllers.CleanupReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		OperatorNamespace: operatorNamespace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CleanupController")
		os.Exit(1)
	}

	if err = (&webhook.ClusterServiceVersionDeploymentScaler{
		Client:            mgr.GetClient(),
		Decoder:           admission.NewDecoder(mgr.GetScheme()),
		OperatorNamespace: operatorNamespace,
	}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterServiceVersion")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func getCacheOptions(operatorNamespace string) (cache.Options, error) {

	// Obtain config (works in-cluster and with KUBECONFIG outside)
	cfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "error getting kubeconfig")
		return cache.Options{}, err
	}

	// Create the client
	cli, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "error creating client")
		return cache.Options{}, err
	}

	configmap, err := controllers.GetOdfConfigMap(context.Background(), cli, setupLog)
	if err != nil {
		setupLog.Error(err, "error getting configmap")
		return cache.Options{}, err
	}

	defaultNamespaces := map[string]cache.Config{
		operatorNamespace:            {},
		"openshift-storage-extended": {},
	}

	relevantCRDNames := map[string]bool{}

	controllers.ParseOdfConfigMapRecords(setupLog, configmap, func(record *controllers.OdfOperatorConfigMapRecord, key, rawValue string) {

		for _, crdName := range record.ScaleUpOnInstanceOf {
			relevantCRDNames[crdName] = true
		}
	})

	cacheOptions := cache.Options{
		DefaultNamespaces: defaultNamespaces,
		// Set cache for relevant CRDs only
		ByObject: map[client.Object]cache.ByObject{
			&extv1.CustomResourceDefinition{}: {
				Transform: func(obj any) (any, error) {
					if crd, ok := obj.(*extv1.CustomResourceDefinition); ok {
						if ok := relevantCRDNames[crd.Name]; ok {
							return crd, nil // keep in cache
						}
					}
					return nil, nil // drop from cache
				},
			},
		},
	}

	return cacheOptions, nil
}
