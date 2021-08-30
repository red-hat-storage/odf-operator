# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 4.9.0

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMAGE_REGISTRY ?= quay.io
REGISTRY_NAMESPACE ?= ocs-dev
IMAGE_TAG ?= latest
IMAGE_NAME ?= odf-operator
BUNDLE_IMAGE_NAME ?= $(IMAGE_NAME)-bundle
CATALOG_IMAGE_NAME ?= $(IMAGE_NAME)-catalog

# IMG defines the image used for the operator.
IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)

# BUNDLE_IMG defines the image used for the bundle.
BUNDLE_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(BUNDLE_IMAGE_NAME):$(IMAGE_TAG)

# CATALOG_IMG defines the image used for the catalog.
CATALOG_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(CATALOG_IMAGE_NAME):$(IMAGE_TAG)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

OCS_BUNDLE_NAME ?= ocs-operator
OCS_BUNDLE_IMG_NAME ?= $(OCS_BUNDLE_NAME)-bundle
OCS_BUNDLE_IMG_TAG ?= v4.9.0
OCS_BUNDLE_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(OCS_BUNDLE_IMG_NAME):$(OCS_BUNDLE_IMG_TAG)

IBM_BUNDLE_NAME ?= ibm-storage-odf-operator
IBM_BUNDLE_IMG_NAME ?= $(IBM_BUNDLE_NAME)-bundle
IBM_BUNDLE_IMG_TAG ?= 0.2.0
IBM_BUNDLE_IMG ?= docker.io/ibmcom/$(IBM_BUNDLE_IMG_NAME):$(IBM_BUNDLE_IMG_TAG)

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(shell echo $(BUNDLE_IMG) $(OCS_BUNDLE_IMG) $(IBM_BUNDLE_IMG) | sed "s/ /,/g")

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# manager env variables
OCS_CSV_NAME ?= $(OCS_BUNDLE_NAME).$(OCS_BUNDLE_IMG_TAG)
IBM_SUBSCRIPTION_NAME ?= $(IBM_BUNDLE_NAME)
IBM_SUBSCRIPTION_PACKAGE ?= $(IBM_BUNDLE_NAME)
IBM_SUBSCRIPTION_CHANNEL ?= stable-v1
IBM_SUBSCRIPTION_STARTINGCSV ?= $(IBM_BUNDLE_NAME).v$(IBM_BUNDLE_IMG_TAG)
IBM_SUBSCRIPTION_CATALOGSOURCE ?= odf-catalogsource
IBM_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE ?= openshift-marketplace

# kube rbac proxy image variables
CLUSTER_ENV ?= openshift
KUBE_RBAC_PROXY_IMG ?= gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
OSE_KUBE_RBAC_PROXY_IMG ?= registry.redhat.io/openshift4/ose-kube-rbac-proxy:v4.7.0

ifeq ($(CLUSTER_ENV), openshift)
	RBAC_PROXY_IMG ?= $(OSE_KUBE_RBAC_PROXY_IMG)
else ifeq ($(CLUSTER_ENV), kubernetes)
	RBAC_PROXY_IMG ?= $(KUBE_RBAC_PROXY_IMG)
endif

ODF_CONSOLE_IMG_NAME ?= odf-console
ODF_CONSOLE_IMG_TAG ?= latest
ODF_CONSOLE_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(ODF_CONSOLE_IMG_NAME):$(ODF_CONSOLE_IMG_TAG)

IBM_CONSOLE_IMG_NAME ?= ibm-storage-odf-plugin
IBM_CONSOLE_IMG_TAG ?= 0.2.0
IBM_CONSOLE_IMG ?= docker.io/ibmcom/$(IBM_CONSOLE_IMG_NAME):$(IBM_CONSOLE_IMG_TAG)
