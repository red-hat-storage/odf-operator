package e2e

import (
	"github.com/onsi/gomega"
)

// TestUpgrade tests the upgrade of ODF operator
func UpgradeTest() {

	debug("Upgrade Test: started\n")

	SuiteFailed = true

	debug("Upgrade Test: Uninstalling the current version of ODF\n")
	err := DeployManager.BeforeUpgradeCleanup(OdfCatalogSourceImage, OdfSubscriptionChannel)
	gomega.Expect(err).To(gomega.BeNil())

	debug("Upgrade Test: Installing the upgradefrom version of ODF\n")
	err = DeployManager.BeforeUpgradeSetup(UpgradeFromOdfCatalogSourceImage, UpgradeFromOdfSubscriptionChannel)
	gomega.Expect(err).To(gomega.BeNil())

	debug("Upgrade Test: Upgrading ODF\n")
	err = DeployManager.UpgradeODFwithOLM(OdfCatalogSourceImage, OdfSubscriptionChannel)
	gomega.Expect(err).To(gomega.BeNil())

	debug("Upgrade Test: Waiting for Upgraded CSVs\n")
	err = DeployManager.CheckAllCsvs(CsvNames)
	gomega.Expect(err).To(gomega.BeNil())

	debug("Upgrade Test: Checking the Storage System\n")
	err = DeployManager.CheckStorageSystem()
	gomega.Expect(err).To(gomega.BeNil())

	debug("Upgrade Test: Cleanup resources after upgrade\n")
	err = DeployManager.AfterUpgradeCleanup()
	gomega.Expect(err).To(gomega.BeNil())

	SuiteFailed = false

	debug("Upgrade Test: completed\n")
}
