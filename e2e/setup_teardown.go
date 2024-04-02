package e2e

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// Setup is the function called to initialize the test environment
func Setup() {

	debug("Setup: started\n")

	SuiteFailed = true

	if OdfOperatorInstall {
		debug("Setup: deploying ODF Operator\n")
		err := DeployManager.DeployODFWithOLM(OdfCatalogSourceImage, OdfSubscriptionChannel)
		gomega.Expect(err).To(gomega.BeNil())
	}

	debug("Setup: Checking if all the CSVs have succeeded\n")
	err := DeployManager.CheckAllCsvs(CsvNames)
	gomega.Expect(err).To(gomega.BeNil())

	SuiteFailed = false

	debug("Setup: completed\n")
}

// TearDown is the function called to tear down the test environment
func TearDown() {

	debug("TearDown: started\n")

	SuiteFailed = true

	if OdfClusterUninstall {
		debug("TearDown: uninstalling ODF Operator\n")
		err := DeployManager.UndeployODFWithOLM(OdfCatalogSourceImage, OdfSubscriptionChannel)
		gomega.Expect(err).To(gomega.BeNil())
	}

	SuiteFailed = false

	debug("TearDown: completed\n")
}

func debug(msg string, args ...interface{}) {
	ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf(msg, args...))) //nolint:errcheck
}
