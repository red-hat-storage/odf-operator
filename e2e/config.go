package e2e

import (
	"flag"
	"fmt"

	"github.com/red-hat-storage/odf-operator/pkg/deploymanager"
)

var (
	// OdfCatalogSourceImage is the CatalogSource container image
	OdfCatalogSourceImage string
	// OdfSubscriptionChannel is the name of the odf subscription channel
	OdfSubscriptionChannel string
	// OdfOperatorInstall indicates to install the operator
	OdfOperatorInstall bool
	// OdfClusterUninstall indicates to uninstall the operator
	OdfClusterUninstall bool
	// OdfClusterServiceVersion is the name of odf csv
	OdfClusterServiceVersion string
	// OcsClusterServiceVersion is the name of ocs csv
	OcsClusterServiceVersion string
	// NoobaClusterServiceVersion is the name of Nooba csv
	NoobaClusterServiceVersion string
	// CsiaddonsClusterServiceVersion is the name of Csiaddons csv
	CsiaddonsClusterServiceVersion string
)

var (
	// DeployManager is the suite global DeployManager
	DeployManager *deploymanager.DeployManager

	// SuiteFailed indicates whether any test in the current suite has failed
	SuiteFailed = false

	// A list of all the csvs that should be installed
	CsvNames []string
)

func init() {
	flag.StringVar(&OdfCatalogSourceImage, "odf-catalog-image", "", "The ODF CatalogSource container image to use in the deployment")
	flag.StringVar(&OdfSubscriptionChannel, "odf-subscription-channel", "", "The subscription channel to receive updates from")
	flag.BoolVar(&OdfOperatorInstall, "odf-operator-install", true, "Install the ODF operator before starting tests")
	flag.BoolVar(&OdfClusterUninstall, "odf-operator-uninstall", true, "Uninstall the ODF operator after test completion")
	flag.StringVar(&OdfClusterServiceVersion, "odf-cluster-service-version", "", "The ODF CSV name which needs to verified")
	flag.StringVar(&OcsClusterServiceVersion, "ocs-cluster-service-version", "", "The OCS CSV name which needs to verified")
	flag.StringVar(&NoobaClusterServiceVersion, "nooba-cluster-service-version", "", "The Nooba CSV name which needs to verified")
	flag.StringVar(&CsiaddonsClusterServiceVersion, "csiaddons-cluster-service-version", "", "The CSI Addon CSV name which needs to verified")
	flag.Parse()

	verifyFlags()

	// A list of names of all the csvs that should be installed
	CsvNames = []string{OdfClusterServiceVersion, OcsClusterServiceVersion, NoobaClusterServiceVersion, CsiaddonsClusterServiceVersion}

	dm, err := deploymanager.NewDeployManager()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize DeployManager: %v", err))
	}

	DeployManager = dm
}

func verifyFlags() {
	if OdfCatalogSourceImage == "" {
		panic("odf-catalog-image is not provided")
	}

	if OdfSubscriptionChannel == "" {
		panic("odf-subscription-channel is not provided")
	}

	if OdfClusterServiceVersion == "" {
		panic("odf-cluster-service-version is not provided")
	}

	if OcsClusterServiceVersion == "" {
		panic("ocs-cluster-service-version is not provided")
	}

	if NoobaClusterServiceVersion == "" {
		panic("nooba-cluster-service-version is not provided")
	}

	if CsiaddonsClusterServiceVersion == "" {
		panic("csiaddons-cluster-service-version is not provided")
	}
}
