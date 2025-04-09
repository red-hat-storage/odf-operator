# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 4.19.0

# MAX_OCP_VERSION variable specifies the maximum supported version of OCP.
# Its purpose is to add an annotation to the CSV file, blocking OCP upgrades beyond the X+1 version.
MAX_OCP_VERSION := $(shell echo $(VERSION) | awk -F. '{print $$1"."$$2+1}')

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
DEFAULT_CHANNEL ?= alpha
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
CHANNELS ?= $(DEFAULT_CHANNEL)
BUNDLE_CHANNELS := --channels=$(CHANNELS)

BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# OPM_RENDER_OPTS will be used while rendering bundle images
OPM_RENDER_OPTS ?=

# Each CSV has a replaces parameter that indicates which Operator it replaces.
# This builds a graph of CSVs that can be queried by OLM, and updates can be
# shared between channels. Channels can be thought of as entry points into
# the graph of updates:
REPLACES ?=

# Creating the New CatalogSource requires publishing CSVs that replace one Operator,
# but can skip several. This can be accomplished using the skipRange annotation:
SKIP_RANGE ?=

# Image URL to use all building/pushing image targets
IMAGE_REGISTRY ?= quay.io
REGISTRY_NAMESPACE ?= ocs-dev
IMAGE_TAG ?= latest
IMAGE_NAME ?= odf-operator
BUNDLE_IMAGE_NAME ?= $(IMAGE_NAME)-bundle
ODF_DEPS_BUNDLE_NAME ?= odf-dependencies
ODF_DEPS_BUNDLE_IMAGE_NAME ?= $(ODF_DEPS_BUNDLE_NAME)-bundle
CATALOG_IMAGE_NAME ?= $(IMAGE_NAME)-catalog
ODF_DEPS_CATALOG_IMAGE_NAME ?= $(ODF_DEPS_BUNDLE_NAME)-catalog

# IMG defines the image used for the operator.
IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)

# BUNDLE_IMG defines the image used for the bundle.
BUNDLE_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(BUNDLE_IMAGE_NAME):$(IMAGE_TAG)

# ODF_DEPS_BUNDLE_IMG defines the image used for the odf-dependencies bundle.
ODF_DEPS_BUNDLE_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(ODF_DEPS_BUNDLE_IMAGE_NAME):$(IMAGE_TAG)

# CATALOG_IMG defines the image used for the catalog.
CATALOG_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(CATALOG_IMAGE_NAME):$(IMAGE_TAG)

# ODF_DEPS_CATALOG_IMG defines the image used for the deps catalog.
ODF_DEPS_CATALOG_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(ODF_DEPS_CATALOG_IMAGE_NAME):$(IMAGE_TAG)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# manager env variables
OPERATOR_NAMESPACE ?= openshift-storage

ODF_CONSOLE_IMG ?= quay.io/ocs-dev/odf-console:latest

ODF_DEPS_SUBSCRIPTION_PACKAGE ?= $(ODF_DEPS_BUNDLE_NAME)
ODF_DEPS_SUBSCRIPTION_CHANNEL ?= $(DEFAULT_CHANNEL)
ODF_DEPS_SUBSCRIPTION_CSVNAME ?= $(ODF_DEPS_SUBSCRIPTION_PACKAGE).v$(VERSION)

OCS_BUNDLE_IMG ?= quay.io/ocs-dev/ocs-operator-bundle:main-552d231
OCS_BUNDLE_VERSION ?= v4.19.0
OCS_SUBSCRIPTION_PACKAGE ?= ocs-operator
OCS_SUBSCRIPTION_CHANNEL ?= alpha
OCS_SUBSCRIPTION_CSVNAME ?= $(OCS_SUBSCRIPTION_PACKAGE).$(OCS_BUNDLE_VERSION)

OCS_CLIENT_BUNDLE_IMG ?= quay.io/ocs-dev/ocs-client-operator-bundle:3c618ad
OCS_CLIENT_BUNDLE_VERSION ?= v4.19.0
OCS_CLIENT_SUBSCRIPTION_PACKAGE ?= ocs-client-operator
OCS_CLIENT_SUBSCRIPTION_CHANNEL ?= alpha
OCS_CLIENT_SUBSCRIPTION_CSVNAME ?= $(OCS_CLIENT_SUBSCRIPTION_PACKAGE).$(OCS_CLIENT_BUNDLE_VERSION)

ROOK_BUNDLE_IMG ?= quay.io/ocs-dev/rook-ceph-operator-bundle:master-320de2454
ROOK_BUNDLE_VERSION ?= v4.19.0
ROOK_SUBSCRIPTION_PACKAGE ?= rook-ceph-operator
ROOK_SUBSCRIPTION_CHANNEL ?= alpha
ROOK_SUBSCRIPTION_CSVNAME ?= $(ROOK_SUBSCRIPTION_PACKAGE).$(ROOK_BUNDLE_VERSION)

NOOBAA_BUNDLE_IMG ?= quay.io/noobaa/noobaa-operator-bundle:master-20250326
NOOBAA_BUNDLE_VERSION ?= v5.19.0
NOOBAA_SUBSCRIPTION_PACKAGE ?= noobaa-operator
NOOBAA_SUBSCRIPTION_CHANNEL ?= alpha
NOOBAA_SUBSCRIPTION_CSVNAME ?= $(NOOBAA_SUBSCRIPTION_PACKAGE).$(NOOBAA_BUNDLE_VERSION)

CEPHCSI_BUNDLE_IMG ?= quay.io/ocs-dev/cephcsi-operator-bundle:main-0ac7669
CEPHCSI_BUNDLE_VERSION ?= v4.19.0
CEPHCSI_SUBSCRIPTION_PACKAGE ?= cephcsi-operator
CEPHCSI_SUBSCRIPTION_CHANNEL ?= alpha
CEPHCSI_SUBSCRIPTION_CSVNAME ?= $(CEPHCSI_SUBSCRIPTION_PACKAGE).$(CEPHCSI_BUNDLE_VERSION)

CSIADDONS_BUNDLE_IMG ?= quay.io/csiaddons/k8s-bundle:v0.12.0
CSIADDONS_BUNDLE_VERSION ?= v0.12.0
CSIADDONS_SUBSCRIPTION_PACKAGE ?= csi-addons
CSIADDONS_SUBSCRIPTION_CHANNEL ?= alpha
CSIADDONS_SUBSCRIPTION_CSVNAME ?= $(CSIADDONS_SUBSCRIPTION_PACKAGE).$(CSIADDONS_BUNDLE_VERSION)

PROMETHEUS_BUNDLE_IMG ?= quay.io/ocs-dev/odf-prometheus-operator-bundle:main-552d231
PROMETHEUS_BUNDLE_VERSION ?= v4.19.0
PROMETHEUS_SUBSCRIPTION_PACKAGE ?= odf-prometheus-operator
PROMETHEUS_SUBSCRIPTION_CHANNEL ?= beta
PROMETHEUS_SUBSCRIPTION_CSVNAME ?= $(PROMETHEUS_SUBSCRIPTION_PACKAGE).$(PROMETHEUS_BUNDLE_VERSION)

RECIPE_BUNDLE_IMG ?= quay.io/ramendr/recipe-bundle:latest
RECIPE_BUNDLE_VERSION ?= v0.0.1
RECIPE_SUBSCRIPTION_PACKAGE ?= recipe
RECIPE_SUBSCRIPTION_CHANNEL ?= alpha
RECIPE_SUBSCRIPTION_CSVNAME ?= $(RECIPE_SUBSCRIPTION_PACKAGE).$(RECIPE_BUNDLE_VERSION)

IBM_ODF_BUNDLE_IMG ?= quay.io/ibmodffs/ibm-storage-odf-operator-bundle:1.7.0
IBM_ODF_BUNDLE_VERSION ?= v1.7.0
IBM_ODF_SUBSCRIPTION_PACKAGE ?= ibm-storage-odf-operator
IBM_ODF_SUBSCRIPTION_CHANNEL ?= stable-v1.7
IBM_ODF_SUBSCRIPTION_CSVNAME ?= $(IBM_ODF_SUBSCRIPTION_PACKAGE).$(IBM_ODF_BUNDLE_VERSION)


# A space-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0 example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG) \
	$(ODF_DEPS_BUNDLE_IMG) \
	$(OCS_BUNDLE_IMG) \
	$(OCS_CLIENT_BUNDLE_IMG) \
	$(ROOK_BUNDLE_IMG) \
	$(NOOBAA_BUNDLE_IMG) \
	$(CEPHCSI_BUNDLE_IMG) \
	$(CSIADDONS_BUNDLE_IMG) \
	$(PROMETHEUS_BUNDLE_IMG) \
	$(RECIPE_BUNDLE_IMG_TAG) \
	$(IBM_ODF_BUNDLE_IMG)

# The 'odf-operator' CSV name must always be at index 0 in this list,
# as some e2e tests explicitly skip the first element.
CSV_NAMES ?= $(IMAGE_NAME).v$(VERSION) \
	$(ODF_DEPS_SUBSCRIPTION_CSVNAME) \
	$(OCS_SUBSCRIPTION_CSVNAME) \
	$(OCS_CLIENT_SUBSCRIPTION_CSVNAME) \
	$(ROOK_SUBSCRIPTION_CSVNAME) \
	$(NOOBAA_SUBSCRIPTION_CSVNAME) \
	$(CEPHCSI_SUBSCRIPTION_CSVNAME) \
	$(CSIADDONS_SUBSCRIPTION_CSVNAME) \
	$(PROMETHEUS_SUBSCRIPTION_CSVNAME) \
	$(RECIPE_SUBSCRIPTION_CSVNAME)
