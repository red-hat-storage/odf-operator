package odf_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	tests "github.com/red-hat-storage/odf-operator/e2e"
)

func TestOdf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Odf Suite")
}

var _ = BeforeSuite(func() {
	tests.Setup()
})

var _ = AfterSuite(func() {
	tests.TearDown()
})

// Test for the creation & deletion of storagesystem
var _ = Describe("StorageSystem test", func() {
	Context("Checking the StorageSystem", func() {
		It("Should check the creation & deletion of storagesystem", func() {
			tests.TestStorageSystem()
		})
	})
})
