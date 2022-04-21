package odf_test

import (
	"fmt"
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

var _ = Describe("odftest", func() {
	Context("Run a dummy test", func() {
		It("Should do nothing", func() {
			fmt.Println("Just a test")
		})
	})
})
