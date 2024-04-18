# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 4.16.0

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

# Each CSV has the option to specify an annotation 'operators.operatorframework.io/operator-type',
# which is an annotation that is (only!) read by the OLM Console UI to determine the visibility of
# the Operator package/bundle in the Operator Hub UI.
#
# Current supported values are 'standalone' (visible) and 'non-standalone' (not visible)
OPERATOR_TYPE ?= standalone

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
CATALOG_IMAGE_NAME ?= $(IMAGE_NAME)-catalog

# IMG defines the image used for the operator.
IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)

# BUNDLE_IMG defines the image used for the bundle.
BUNDLE_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(BUNDLE_IMAGE_NAME):$(IMAGE_TAG)

# CATALOG_IMG defines the image used for the catalog.
CATALOG_IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(CATALOG_IMAGE_NAME):$(IMAGE_TAG)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

OCS_BUNDLE_NAME ?= ocs-operator
OCS_BUNDLE_IMG_NAME ?= $(OCS_BUNDLE_NAME)-bundle
OCS_BUNDLE_VERSION ?= v4.16.0
OCS_BUNDLE_IMG_TAG ?= release-4.16-723d642
OCS_BUNDLE_IMG_LOCATION ?= quay.io/ocs-dev
OCS_BUNDLE_IMG ?= $(OCS_BUNDLE_IMG_LOCATION)/$(OCS_BUNDLE_IMG_NAME):$(OCS_BUNDLE_IMG_TAG)

OCS_CLIENT_BUNDLE_NAME ?= ocs-client-operator
OCS_CLIENT_BUNDLE_IMG_NAME ?= $(OCS_CLIENT_BUNDLE_NAME)-bundle
OCS_CLIENT_BUNDLE_VERSION ?= v4.16.0
OCS_CLIENT_BUNDLE_IMG_TAG ?= main-a79f241-addons-080
OCS_CLIENT_BUNDLE_IMG_LOCATION ?= quay.io/ocs-dev
OCS_CLIENT_BUNDLE_IMG ?= $(OCS_CLIENT_BUNDLE_IMG_LOCATION)/$(OCS_CLIENT_BUNDLE_IMG_NAME):$(OCS_CLIENT_BUNDLE_IMG_TAG)

NOOBAA_BUNDLE_NAME ?= noobaa-operator
NOOBAA_BUNDLE_IMG_NAME ?= $(NOOBAA_BUNDLE_NAME)-bundle
NOOBAA_BUNDLE_VERSION ?= v5.14.0
NOOBAA_BUNDLE_IMG_TAG ?= v5.14.0
NOOBAA_BUNDLE_IMG_LOCATION ?= quay.io/noobaa
NOOBAA_BUNDLE_IMG ?= $(NOOBAA_BUNDLE_IMG_LOCATION)/$(NOOBAA_BUNDLE_IMG_NAME):$(NOOBAA_BUNDLE_IMG_TAG)

CSIADDONS_BUNDLE_NAME ?= csi-addons
CSIADDONS_BUNDLE_IMG_NAME ?= k8s-bundle
CSIADDONS_BUNDLE_VERSION ?= v0.8.0
CSIADDONS_BUNDLE_IMG_TAG ?= v0.8.0
CSIADDONS_BUNDLE_IMG_LOCATION ?= quay.io/csiaddons
CSIADDONS_BUNDLE_IMG ?= $(CSIADDONS_BUNDLE_IMG_LOCATION)/$(CSIADDONS_BUNDLE_IMG_NAME):$(CSIADDONS_BUNDLE_IMG_TAG)

IBM_BUNDLE_NAME ?= ibm-storage-odf-operator
IBM_BUNDLE_IMG_NAME ?= $(IBM_BUNDLE_NAME)-bundle
IBM_BUNDLE_VERSION ?= 1.4.1
IBM_BUNDLE_IMG_TAG ?= 1.4.1
IBM_BUNDLE_IMG_LOCATION ?= quay.io/ibmodffs
IBM_BUNDLE_IMG ?= $(IBM_BUNDLE_IMG_LOCATION)/$(IBM_BUNDLE_IMG_NAME):$(IBM_BUNDLE_IMG_TAG)

ODF_CONSOLE_IMG_NAME ?= odf-console
ODF_CONSOLE_IMG_TAG ?= latest
ODF_CONSOLE_IMG_LOCATION ?= quay.io/ocs-dev
ODF_CONSOLE_IMG ?= $(ODF_CONSOLE_IMG_LOCATION)/$(ODF_CONSOLE_IMG_NAME):$(ODF_CONSOLE_IMG_TAG)

ROOK_BUNDLE_NAME ?= rook-ceph-operator
ROOK_BUNDLE_IMG_NAME ?= $(ROOK_BUNDLE_NAME)-bundle
ROOK_BUNDLE_VERSION ?= v4.16.0
ROOK_BUNDLE_IMG_TAG ?= master-f8e5f814c
ROOK_BUNDLE_IMG_LOCATION ?= quay.io/ocs-dev
ROOK_BUNDLE_IMG ?= $(ROOK_BUNDLE_IMG_LOCATION)/$(ROOK_BUNDLE_IMG_NAME):$(ROOK_BUNDLE_IMG_TAG)

# To be changed when odf-prometheus-operator bundle exists
PROMETHEUS_BUNDLE_NAME ?= odf-prometheus-operator
PROMETHEUS_BUNDLE_IMG_NAME ?= $(PROMETHEUS_BUNDLE_NAME)-bundle
PROMETHEUS_BUNDLE_VERSION ?= v4.16.0
PROMETHEUS_BUNDLE_IMG_TAG ?= main-d5fbc29
PROMETHEUS_BUNDLE_IMG_LOCATION ?= quay.io/ocs-dev
PROMETHEUS_BUNDLE_IMG ?= $(PROMETHEUS_BUNDLE_IMG_LOCATION)/$(PROMETHEUS_BUNDLE_IMG_NAME):$(PROMETHEUS_BUNDLE_IMG_TAG)

# A space-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0 example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG) $(OCS_BUNDLE_IMG) $(OCS_CLIENT_BUNDLE_IMG) $(IBM_BUNDLE_IMG) $(NOOBAA_BUNDLE_IMG) $(CSIADDONS_BUNDLE_IMG) $(ROOK_BUNDLE_IMG) $(PROMETHEUS_BUNDLE_IMG)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# manager env variables
OPERATOR_NAMESPACE ?= openshift-storage
OPERATOR_CATALOGSOURCE ?= odf-catalogsource
OPERATOR_CATALOGSOURCE_NAMESPACE ?= openshift-marketplace

NOOBAA_SUBSCRIPTION_NAME ?= $(NOOBAA_BUNDLE_NAME)
NOOBAA_SUBSCRIPTION_PACKAGE ?= $(NOOBAA_BUNDLE_NAME)
NOOBAA_SUBSCRIPTION_CHANNEL ?= $(DEFAULT_CHANNEL)
NOOBAA_SUBSCRIPTION_STARTINGCSV ?= $(NOOBAA_BUNDLE_NAME).$(NOOBAA_BUNDLE_VERSION)
NOOBAA_SUBSCRIPTION_CATALOGSOURCE ?= $(OPERATOR_CATALOGSOURCE)
NOOBAA_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE ?= $(OPERATOR_CATALOGSOURCE_NAMESPACE)

CSIADDONS_SUBSCRIPTION_NAME ?= $(CSIADDONS_BUNDLE_NAME)
CSIADDONS_SUBSCRIPTION_PACKAGE ?= $(CSIADDONS_BUNDLE_NAME)
CSIADDONS_SUBSCRIPTION_CHANNEL ?= $(DEFAULT_CHANNEL)
CSIADDONS_SUBSCRIPTION_STARTINGCSV ?= $(CSIADDONS_BUNDLE_NAME).$(CSIADDONS_BUNDLE_VERSION)
CSIADDONS_SUBSCRIPTION_CATALOGSOURCE ?= $(OPERATOR_CATALOGSOURCE)
CSIADDONS_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE ?= $(OPERATOR_CATALOGSOURCE_NAMESPACE)

OCS_SUBSCRIPTION_NAME ?= $(OCS_BUNDLE_NAME)
OCS_SUBSCRIPTION_PACKAGE ?= $(OCS_BUNDLE_NAME)
OCS_SUBSCRIPTION_CHANNEL ?= $(DEFAULT_CHANNEL)
OCS_SUBSCRIPTION_STARTINGCSV ?= $(OCS_BUNDLE_NAME).$(OCS_BUNDLE_VERSION)
OCS_SUBSCRIPTION_CATALOGSOURCE ?= $(OPERATOR_CATALOGSOURCE)
OCS_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE ?= $(OPERATOR_CATALOGSOURCE_NAMESPACE)

OCS_CLIENT_SUBSCRIPTION_NAME ?= $(OCS_CLIENT_BUNDLE_NAME)
OCS_CLIENT_SUBSCRIPTION_PACKAGE ?= $(OCS_CLIENT_BUNDLE_NAME)
OCS_CLIENT_SUBSCRIPTION_CHANNEL ?= $(DEFAULT_CHANNEL)
OCS_CLIENT_SUBSCRIPTION_STARTINGCSV ?= $(OCS_CLIENT_BUNDLE_NAME).$(OCS_CLIENT_BUNDLE_VERSION)
OCS_CLIENT_SUBSCRIPTION_CATALOGSOURCE ?= $(OPERATOR_CATALOGSOURCE)
OCS_CLIENT_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE ?= $(OPERATOR_CATALOGSOURCE_NAMESPACE)

IBM_SUBSCRIPTION_NAME ?= $(IBM_BUNDLE_NAME)
IBM_SUBSCRIPTION_PACKAGE ?= $(IBM_BUNDLE_NAME)
IBM_SUBSCRIPTION_CHANNEL ?= stable-v1.4
IBM_SUBSCRIPTION_STARTINGCSV ?= $(IBM_BUNDLE_NAME).v$(IBM_BUNDLE_VERSION)
IBM_SUBSCRIPTION_CATALOGSOURCE ?= $(OPERATOR_CATALOGSOURCE)
IBM_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE ?= $(OPERATOR_CATALOGSOURCE_NAMESPACE)

ROOK_SUBSCRIPTION_NAME ?= $(ROOK_BUNDLE_NAME)
ROOK_SUBSCRIPTION_PACKAGE ?= $(ROOK_BUNDLE_NAME)
ROOK_SUBSCRIPTION_CHANNEL ?= $(DEFAULT_CHANNEL)
ROOK_SUBSCRIPTION_STARTINGCSV ?= $(ROOK_BUNDLE_NAME).$(ROOK_BUNDLE_VERSION)
ROOK_SUBSCRIPTION_CATALOGSOURCE ?= $(OPERATOR_CATALOGSOURCE)
ROOK_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE ?= $(OPERATOR_CATALOGSOURCE_NAMESPACE)

PROMETHEUS_SUBSCRIPTION_NAME ?= $(PROMETHEUS_BUNDLE_NAME)
PROMETHEUS_SUBSCRIPTION_PACKAGE ?= $(PROMETHEUS_BUNDLE_NAME)
PROMETHEUS_SUBSCRIPTION_CHANNEL ?= beta
PROMETHEUS_SUBSCRIPTION_STARTINGCSV ?= $(PROMETHEUS_BUNDLE_NAME).$(PROMETHEUS_BUNDLE_VERSION)
PROMETHEUS_SUBSCRIPTION_CATALOGSOURCE ?= $(OPERATOR_CATALOGSOURCE)
PROMETHEUS_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE ?= $(OPERATOR_CATALOGSOURCE_NAMESPACE)
# kube rbac proxy image variables
CLUSTER_ENV ?= openshift
KUBE_RBAC_PROXY_IMG ?= gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1
OSE_KUBE_RBAC_PROXY_IMG ?= registry.redhat.io/openshift4/ose-kube-rbac-proxy:v4.11.0

ifeq ($(CLUSTER_ENV), openshift)
	RBAC_PROXY_IMG ?= $(OSE_KUBE_RBAC_PROXY_IMG)
else ifeq ($(CLUSTER_ENV), kubernetes)
	RBAC_PROXY_IMG ?= $(KUBE_RBAC_PROXY_IMG)
endif
