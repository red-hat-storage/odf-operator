package deploymanager

import (
	stderrors "errors"
	"fmt"

	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"go.uber.org/multierr"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/red-hat-storage/odf-operator/controllers"
)

var ErrDeploymentsNotScaledUp = stderrors.New("deployments are not scaled up")

func (d *DeployManager) ValidateOperatorScaler() error {

	var kindMapping = map[string]*controllers.KindCsvsRecord{}
	if err := d.LoadOdfConfigMapData(kindMapping); err != nil {
		d.Log.Error(err, "failed to load configmap")
		return err
	}

	for _, kindCsvRecord := range kindMapping {

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": kindCsvRecord.ApiVersion,
				"kind":       kindCsvRecord.Kind,
				"metadata": map[string]any{
					"name":      "test-obj",
					"namespace": InstallNamespace,
				},
				// Add required spec fields here
				"spec": map[string]any{
					// ... required fields for your CRD
				},
			},
		}

		if kindCsvRecord.Kind == "NooBaa" {
			// Set the AllowNoobaaDeletion field
			if err := unstructured.SetNestedField(
				obj.Object, true,
				"spec", "cleanupPolicy", "allowNoobaaDeletion"); err != nil {
				d.Log.Error(err, "failed to set noobaa cleanup policy")
				return err
			}
		}

		if err := d.Client.Create(d.Ctx, obj); err != nil {
			d.Log.Error(err, "failed to create object", "kind", kindCsvRecord.Kind)
			return err
		}

		// Cleanup: Restore CSV replica to original state
		//nolint:errcheck
		defer d.ScaleDownCsvsDeploymentsReplicas(kindCsvRecord.CsvNames)
		d.Log.Info("cleanup", "csvs", kindCsvRecord.CsvNames)
		// Cleanup: Delete the CR
		//nolint:errcheck
		defer d.Client.Delete(d.Ctx, obj)
		d.Log.Info("cleanup", "kind", kindCsvRecord.Kind)

		if err := retry.OnError(
			retry.DefaultRetry,
			func(err error) bool {
				return stderrors.Is(err, ErrDeploymentsNotScaledUp)
			},
			func() error {
				if err := d.ValidateCsvsDeploymentsReplicasAreScaledUp(kindCsvRecord.CsvNames); err != nil {
					d.Log.Error(err, "failed to validate csv deployment replicas")
					return err
				}
				return nil
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func (d *DeployManager) LoadOdfConfigMapData(kindMapping map[string]*controllers.KindCsvsRecord) error {

	configmap, err := controllers.GetOdfConfigMap(d.Ctx, d.Client, d.Log)
	if err != nil {
		d.Log.Error(err, "failed getting odf configmap")
		return err
	}

	var combinedErr error

	controllers.ParseOdfConfigMapRecords(d.Log, configmap, func(record *controllers.OdfOperatorConfigMapRecord, key, rawValue string) {

		if record.Csv == "" || record.ScaleUpOnInstanceOf == nil {
			d.Log.Info("skipping the record from the configmap", "key", key, "value", rawValue)
			return
		}
		for _, crdName := range record.ScaleUpOnInstanceOf {

			rec, ok := kindMapping[crdName]
			if !ok {
				rec = &controllers.KindCsvsRecord{}
			}
			rec.CsvNames = append(rec.CsvNames, record.Csv)

			// populate the apiVersion and kind
			crd := &extv1.CustomResourceDefinition{}
			crd.Name = crdName
			if err := d.Client.Get(d.Ctx, client.ObjectKeyFromObject(crd), crd); errors.IsNotFound(err) {
				d.Log.Info("skipping crd not found", "crdName", crdName)
				continue
			} else if err != nil {
				d.Log.Error(err, "failed getting crd", "crdName", crdName)
				multierr.AppendInto(&combinedErr, err)
				continue
			}

			rec.ApiVersion = crd.Spec.Group + "/" + crd.Spec.Versions[0].Name
			rec.Kind = crd.Spec.Names.Kind
			kindMapping[crdName] = rec
		}
	})

	d.Log.Info("operator scaler records", "records", kindMapping)

	return combinedErr
}

func (d *DeployManager) ValidateCsvsDeploymentsReplicasAreScaledUp(csvNames []string) error {

	for _, csvName := range csvNames {
		csv := &opv1a1.ClusterServiceVersion{}
		if err := d.Client.Get(d.Ctx, client.ObjectKey{Name: csvName, Namespace: InstallNamespace}, csv); err != nil {
			d.Log.Error(err, "failed to get the csv")
			return err
		}

		// Check if the CSV is scaled down
		for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
			deploymentSpec := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec
			if deploymentSpec.Replicas != nil && *deploymentSpec.Replicas == 0 {
				err := fmt.Errorf("csv %s %w", csvName, ErrDeploymentsNotScaledUp)
				d.Log.Error(err, "csv deployments are not scaled up")
				return err
			}
		}
	}

	return nil
}

func (d *DeployManager) ScaleDownCsvsDeploymentsReplicas(csvNames []string) error {

	for _, csvName := range csvNames {
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {

			csv := &opv1a1.ClusterServiceVersion{}
			if err := d.Client.Get(d.Ctx, client.ObjectKey{Name: csvName, Namespace: InstallNamespace}, csv); err != nil {
				d.Log.Error(err, "failed to get the csv")
				return err
			}

			// Scaled down deployments replicas
			for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
				deploymentSpec := &csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec
				deploymentSpec.Replicas = ptr.To(int32(0))
			}

			if err := d.Client.Update(d.Ctx, csv); err != nil {
				d.Log.Error(err, "failed to update the csv")
				return err
			}

			d.Log.Info("successfully scaled down deployments replicas of csv", "csvName", csvName)
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
