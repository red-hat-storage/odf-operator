package e2e

import "github.com/onsi/gomega"

// TestStorageSystem tests the creation & deletion of storagesystem
func TestStorageSystem() {

	debug("Storagesystem Test: started\n")

	SuiteFailed = true

	debug("Storagesystem Test: Creating Storage System\n")
	err := DeployManager.CreateStorageSystem()
	gomega.Expect(err).To(gomega.BeNil())

	debug("Storagesystem Test: Checking Storage System Conditions\n")
	err = DeployManager.CheckStorageSystemCondition()
	gomega.Expect(err).To(gomega.BeNil())

	debug("Storagesystem Test: Checking Storage System Owner References\n")
	err = DeployManager.CheckStorageSystemOwnerRef()
	gomega.Expect(err).To(gomega.BeNil())	

	debug("Storagesystem Test: Deleting Storage System\n")
	err = DeployManager.DeleteStorageSystemAndWait()
	gomega.Expect(err).To(gomega.BeNil())

	SuiteFailed = false

	debug("Storagesystem Test: completed\n")
}