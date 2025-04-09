package e2e

import (
	"flag"
	"fmt"
	"strings"

	"github.com/red-hat-storage/odf-operator/pkg/deploymanager"
)

var (
	// OdfOperatorInstall indicates to install the operator
	OdfOperatorInstall bool
	// OdfClusterUninstall indicates to uninstall the operator
	OdfClusterUninstall bool
	// OdfCatalogSourceImage is the CatalogSource container image
	OdfCatalogSourceImage string
	// OdfSubscriptionChannel is the name of the odf subscription channel
	OdfSubscriptionChannel string
	// OdfClusterServiceVersion is the name of odf csv
	OdfClusterServiceVersion string
	// A list of all the csvs that should be installed
	CsvNames []string
)

var (
	// DeployManager is the suite global DeployManager
	DeployManager *deploymanager.DeployManager

	// SuiteFailed indicates whether any test in the current suite has failed
	SuiteFailed = false
)

func init() {
	flag.BoolVar(&OdfOperatorInstall, "odf-operator-install", true, "Install the ODF operator before starting tests")
	flag.BoolVar(&OdfClusterUninstall, "odf-operator-uninstall", true, "Uninstall the ODF operator after test completion")
	flag.StringVar(&OdfCatalogSourceImage, "odf-catalog-image", "", "The ODF CatalogSource container image to use in the deployment")
	flag.StringVar(&OdfSubscriptionChannel, "odf-subscription-channel", "", "The subscription channel to receive updates from")
	flag.StringVar(&OdfClusterServiceVersion, "odf-cluster-service-version", "", "The ODF CSV name which needs to verified")
	flag.Func("csv-names", "Space-separated list of ODF CSV names to verify", func(s string) error {
		CsvNames = strings.Split(s, " ")
		return nil
	})
	flag.Parse()
	verifyFlags()

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

	if len(CsvNames) == 0 {
		panic("csv-names is not provided")
	}
}
