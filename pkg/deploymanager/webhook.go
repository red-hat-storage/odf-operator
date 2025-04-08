package deploymanager

import (
	"fmt"

	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	admrv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (d *DeployManager) ValidateWebhookResources() error {

	// Check if the Service is present
	service := &corev1.Service{}
	err := d.Client.Get(d.Ctx, client.ObjectKey{Name: "odf-operator-webhook-server-service", Namespace: InstallNamespace}, service)
	if err != nil {
		return err
	}

	// Check if the MutatingWebhookConfiguration is present
	webhook := &admrv1.MutatingWebhookConfiguration{}
	err = d.Client.Get(d.Ctx, client.ObjectKey{Name: "csv.odf.openshift.io"}, webhook)
	if err != nil {
		return err
	}

	return nil
}

func (d *DeployManager) ValidateCsvsDeploymentsReplicasAreScaledDown(csvNames []string) error {

	for _, csvName := range csvNames {
		csv := &opv1a1.ClusterServiceVersion{}
		err := d.Client.Get(d.Ctx, client.ObjectKey{Name: csvName, Namespace: InstallNamespace}, csv)
		if err != nil {
			return err
		}

		// Check if the CSV is scaled down
		for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
			deploymentSpec := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec
			if deploymentSpec.Replicas == nil || *deploymentSpec.Replicas != 0 {
				return fmt.Errorf("csv %s is not scaled down", csvName)
			}
		}
	}

	return nil
}
