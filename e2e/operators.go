package e2e

import "github.com/onsi/gomega"

func TestWebhook() {

	debug("Webhhok Test: started\n")

	SuiteFailed = true

	debug("Webhhok Test: Check Webhook Presence\n")
	err := DeployManager.ValidateWebhookResources()
	gomega.Expect(err).To(gomega.BeNil())

	debug("Webhook Test: Check CSVs Deployment Replicas Scaled Down\n")
	// pass all csv names except odf-operator, because odf-operator will be always scaled up
	err = DeployManager.ValidateCsvsDeploymentsReplicasAreScaledDown(CsvNames[1:])
	gomega.Expect(err).To(gomega.BeNil())

	SuiteFailed = false

	debug("Webhook Test: completed\n")
}

func TestOperatorScaler() {

	debug("OperatorScaler Test: started\n")

	SuiteFailed = true

	err := DeployManager.ValidateOperatorScaler()
	gomega.Expect(err).To(gomega.BeNil())

	SuiteFailed = false

	debug("OperatorScaler Test: completed\n")
}
