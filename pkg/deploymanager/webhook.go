package deploymanager

import (
	stderrors "errors"
	"fmt"

	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	admrv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrDeploymentsNotScaledDown = stderrors.New("deployments are not scaled down")

func (d *DeployManager) ValidateWebhookResources() error {

	// Check if the Service is present
	service := &corev1.Service{}
	if err := d.Client.Get(d.Ctx, client.ObjectKey{Name: "odf-operator-webhook-server-service", Namespace: InstallNamespace}, service); err != nil {
		d.Log.Error(err, "failed to get service")
		return err
	}

	// Check if the MutatingWebhookConfiguration is present
	webhook := &admrv1.MutatingWebhookConfiguration{}
	if err := d.Client.Get(d.Ctx, client.ObjectKey{Name: "csv.odf.openshift.io"}, webhook); err != nil {
		d.Log.Error(err, "failed to get webhook")
		return err
	}

	return nil
}

func (d *DeployManager) ValidateCsvsDeploymentsReplicasAreScaledDown(csvNames []string) error {

	for _, csvName := range csvNames {
		csv := &opv1a1.ClusterServiceVersion{}
		if err := d.Client.Get(d.Ctx, client.ObjectKey{Name: csvName, Namespace: InstallNamespace}, csv); err != nil {
			d.Log.Error(err, "failed to get the csv")
			return err
		}

		// Check if the CSV is scaled down
		for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
			deploymentSpec := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec
			if deploymentSpec.Replicas == nil || *deploymentSpec.Replicas != 0 {
				err := fmt.Errorf("csv %s %w", csvName, ErrDeploymentsNotScaledDown)
				d.Log.Error(err, "csv deployments are not scaled down")
				return err
			}
		}
	}

	return nil
}
