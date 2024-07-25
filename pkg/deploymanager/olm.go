package deploymanager

import (
	"context"
	"fmt"
	"time"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OlmResources struct {
	operatorGroups []*operatorsv1.OperatorGroup
	catalogSources []*operatorsv1alpha1.CatalogSource
	subscriptions  []*operatorsv1alpha1.Subscription
}

// DeployODFWithOLM deploys odf operator via an olm subscription
func (d *DeployManager) DeployODFWithOLM(odfCatalogImage, subscriptionChannel string) error {

	err := d.CreateNamespace(InstallNamespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	olmResources := d.GetOlmResources(odfCatalogImage, subscriptionChannel)
	err = d.CreateOlmResources(olmResources)
	if err != nil {
		return err
	}

	return nil
}

// CheckAllCsvs checks if all the required csvs are present & have succeeded
func (d *DeployManager) CheckAllCsvs(csvNames []string) error {
	for _, csvName := range csvNames {
		csv := &operatorsv1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      csvName,
				Namespace: InstallNamespace,
			},
		}
		err := d.WaitForCsv(csv)
		if err != nil {
			return err
		}
	}
	return nil
}

// UndeployODFWithOLM uninstalls odf operator
func (d *DeployManager) UndeployODFWithOLM(odfCatalogImage, subscriptionChannel string) error {

	olmResources := d.GetOlmResources(odfCatalogImage, subscriptionChannel)
	err := d.DeleteOlmResources(olmResources)
	if err != nil {
		return err
	}

	err = d.DeleteNamespaceAndWait(InstallNamespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// GetOlmResources returns OLM resources required to deploy odf operator
func (d *DeployManager) GetOlmResources(odfCatalogImage, subscriptionChannel string) *OlmResources {

	olmResources := &OlmResources{}

	// Operator Groups
	odfOperatorGroups := &operatorsv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openshift-storage-operatorgroup",
			Namespace: InstallNamespace,
		},
		Spec: operatorsv1.OperatorGroupSpec{
			TargetNamespaces: []string{InstallNamespace},
		},
	}
	olmResources.operatorGroups = append(olmResources.operatorGroups, odfOperatorGroups)

	// Catalog Source
	odfCatalog := &operatorsv1alpha1.CatalogSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "odf-catalogsource",
			Namespace: InstallNamespace,
		},
		Spec: operatorsv1alpha1.CatalogSourceSpec{
			SourceType:  operatorsv1alpha1.SourceTypeGrpc,
			Image:       odfCatalogImage,
			DisplayName: "OpenShift Data Foundation",
			Publisher:   "Red Hat",
		},
	}
	olmResources.catalogSources = append(olmResources.catalogSources, odfCatalog)

	// Subscriptions
	odfSubscription := &operatorsv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "odf-subscription",
			Namespace: InstallNamespace,
		},
		Spec: &operatorsv1alpha1.SubscriptionSpec{
			Channel:                subscriptionChannel,
			InstallPlanApproval:    "Automatic",
			Package:                "odf-operator",
			CatalogSource:          "odf-catalogsource",
			CatalogSourceNamespace: InstallNamespace,
		},
	}
	olmResources.subscriptions = append(olmResources.subscriptions, odfSubscription)

	return olmResources
}

// CreateOlmResources create OLM resources required to deploy odf operator
func (d *DeployManager) CreateOlmResources(olmResources *OlmResources) error {

	for _, operatorGroup := range olmResources.operatorGroups {
		err := d.Client.Create(d.Ctx, operatorGroup)
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}

	for _, catalogSource := range olmResources.catalogSources {
		err := d.Client.Create(d.Ctx, catalogSource)
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
		err = d.WaitForCatalogSource(catalogSource)
		if err != nil {
			return err
		}
	}
	for _, subscription := range olmResources.subscriptions {
		err := d.Client.Create(d.Ctx, subscription)
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

// DeleteOlmResources delete OLM resources required to deploy odf operator
func (d *DeployManager) DeleteOlmResources(olmResources *OlmResources) error {

	for _, subscription := range olmResources.subscriptions {
		err := d.Client.Delete(d.Ctx, subscription)
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}

	for _, catalogSource := range olmResources.catalogSources {
		err := d.Client.Delete(d.Ctx, catalogSource)
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}

	for _, operatorGroup := range olmResources.operatorGroups {
		err := d.Client.Delete(d.Ctx, operatorGroup)
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

// WaitForCatalogSource wait for catalogSource to become ready
func (d *DeployManager) WaitForCatalogSource(catalogsource *operatorsv1alpha1.CatalogSource) error {

	timeout := 600 * time.Second
	interval := 10 * time.Second

	lastReason := ""

	err := utilwait.PollUntilContextTimeout(d.Ctx, interval, timeout, true, func(context.Context) (done bool, err error) {
		existingCatalogSource := &operatorsv1alpha1.CatalogSource{}
		err = d.Client.Get(d.Ctx, client.ObjectKeyFromObject(catalogsource), existingCatalogSource)
		if err != nil {
			lastReason = fmt.Sprintf("failed to get catalogsource: %v", err)
			return false, nil
		}
		if existingCatalogSource.Status.GRPCConnectionState == nil {
			lastReason = "catalogsource connection state is nil"
			return false, nil
		}
		if existingCatalogSource.Status.GRPCConnectionState.LastObservedState != "READY" {
			lastReason = fmt.Sprintf("waiting for catalog source to reach ready state, but stuck in %s state",
				existingCatalogSource.Status.GRPCConnectionState.LastObservedState)
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return fmt.Errorf(lastReason)
	}

	return nil
}

// WaitForCsv waits for the CSV to successfully installed
func (d *DeployManager) WaitForCsv(csv *operatorsv1alpha1.ClusterServiceVersion) error {

	timeout := 600 * time.Second
	interval := 10 * time.Second

	lastReason := ""

	err := utilwait.PollUntilContextTimeout(d.Ctx, interval, timeout, true, func(context.Context) (done bool, err error) {
		existingcsv := &operatorsv1alpha1.ClusterServiceVersion{}
		err = d.Client.Get(d.Ctx, client.ObjectKeyFromObject(csv), existingcsv)
		if err != nil {
			lastReason = fmt.Sprintf("failed to get CSV: %v", err)
			return false, nil
		}
		if existingcsv.Status.Phase != operatorsv1alpha1.CSVPhaseSucceeded {
			lastReason = fmt.Sprintf("waiting for CSV to succeed, but stuck in %s phase", existingcsv.Status.Phase)
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return fmt.Errorf(lastReason)
	}

	return nil
}
