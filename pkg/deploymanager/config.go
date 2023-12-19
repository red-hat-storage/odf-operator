package deploymanager

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(k8sscheme.AddToScheme(scheme))
	utilruntime.Must(odfv1alpha1.AddToScheme(scheme))
	utilruntime.Must(ocsv1.AddToScheme(scheme))
	utilruntime.Must(operatorv1.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
}

type DeployManager struct {
	Client client.Client
	Ctx    context.Context
	Log    logr.Logger
}

// NewDeployManager creates a DeployManager struct with default configuration
func NewDeployManager() (*DeployManager, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		return nil, fmt.Errorf("no KUBECONFIG environment variable set")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return &DeployManager{
		Client: client,
		Log:    log.Log.WithName("DeployManager"),
		Ctx:    context.TODO(),
	}, nil
}
