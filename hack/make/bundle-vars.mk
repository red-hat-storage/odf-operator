OCS_OPERATOR_BUNDLE_IMG ?= quay.io/ocs-dev/ocs-operator-bundle:main-9a87de4
OCS_OPERATOR_PKG_NAME ?= ocs-operator
OCS_OPERATOR_PKG_VERSION ?= 4.22.0
OCS_OPERATOR_PKG_CHANNEL ?= alpha

ROOK_CEPH_BUNDLE_IMG ?= quay.io/ocs-dev/rook-ceph-operator-bundle:master-89eb70f42
ROOK_CEPH_PKG_NAME ?= rook-ceph-operator
ROOK_CEPH_PKG_VERSION ?= 4.22.0
ROOK_CEPH_PKG_CHANNEL ?= alpha

NOOBAA_BUNDLE_IMG ?= quay.io/noobaa/noobaa-operator-bundle:master-20260401
NOOBAA_PKG_NAME ?= noobaa-operator
NOOBAA_PKG_VERSION ?= 5.22.0
NOOBAA_PKG_CHANNEL ?= alpha

OCS_CLIENT_BUNDLE_IMG ?= quay.io/ocs-dev/ocs-client-operator-bundle:main-408b4fb
OCS_CLIENT_PKG_NAME ?= ocs-client-operator
OCS_CLIENT_PKG_VERSION ?= 4.22.0
OCS_CLIENT_PKG_CHANNEL ?= alpha

CEPHCSI_BUNDLE_IMG ?= quay.io/ocs-dev/cephcsi-operator-bundle:main-9bd2093
CEPHCSI_PKG_NAME ?= cephcsi-operator
CEPHCSI_PKG_VERSION ?= 4.22.0
CEPHCSI_PKG_CHANNEL ?= alpha

CSIADDONS_BUNDLE_IMG ?= quay.io/csiaddons/k8s-bundle:v0.14.0
CSIADDONS_PKG_NAME ?= csi-addons
CSIADDONS_PKG_VERSION ?= 0.14.0
CSIADDONS_PKG_CHANNEL ?= alpha

ODF_SNAPSHOT_BUNDLE_IMG ?= quay.io/ocs-dev/snapshot-controller-bundle:main-66bfe3a
ODF_SNAPSHOT_PKG_NAME ?= odf-external-snapshotter-operator
ODF_SNAPSHOT_PKG_VERSION ?= 4.22.0
ODF_SNAPSHOT_PKG_CHANNEL ?= alpha

PROMETHEUS_BUNDLE_IMG ?= quay.io/ocs-dev/odf-prometheus-operator-bundle:main-2a77acf
PROMETHEUS_PKG_NAME ?= odf-prometheus-operator
PROMETHEUS_PKG_VERSION ?= 4.22.0
PROMETHEUS_PKG_CHANNEL ?= beta

OCS_TLS_BUNDLE_IMG ?= quay.io/ocs-dev/ocs-tls-profiles-bundle:main-9fd1952
OCS_TLS_PKG_NAME ?= ocs-tls-profiles
OCS_TLS_PKG_VERSION ?= 4.22.0
OCS_TLS_PKG_CHANNEL ?= alpha

RECIPE_BUNDLE_IMG ?= quay.io/ramendr/recipe-bundle:latest
RECIPE_PKG_NAME ?= recipe
RECIPE_PKG_VERSION ?= 0.0.1
RECIPE_PKG_CHANNEL ?= alpha

IBM_ODF_BUNDLE_IMG ?= quay.io/ibmodffs/ibm-storage-odf-operator-bundle:1.9.0
IBM_ODF_PKG_NAME ?= ibm-storage-odf-operator
IBM_ODF_PKG_VERSION ?= 1.9.0
IBM_ODF_PKG_CHANNEL ?= stable-v1.9

IBM_CSI_BUNDLE_IMG ?= quay.io/ibmcsiblock/ibm-block-csi-operator-bundle:1.13.2
IBM_CSI_PKG_NAME ?= ibm-block-csi-operator
IBM_CSI_PKG_VERSION ?= 1.13.2
IBM_CSI_PKG_CHANNEL ?= stable-v1.13.2

CNSA_BUNDLE_IMG ?= preprod.icr.io/cpopen/ibm-spectrum-scale-operator-bundle:v6.0.1.1_latest
CNSA_PKG_NAME ?= ibm-spectrum-scale-operator
CNSA_PKG_VERSION ?= 60.1.100
CNSA_PKG_CHANNEL ?= stable-v60.1
