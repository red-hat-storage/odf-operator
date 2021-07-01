module github.com/red-hat-data-services/odf-operator

go 1.15

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/openshift/api v3.9.0+incompatible
	github.com/openshift/custom-resource-status v1.1.0
	github.com/operator-framework/api v0.9.2
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/openshift/api => github.com/openshift/api v0.0.0-20201203102015-275406142edb // required for Quickstart CRD
