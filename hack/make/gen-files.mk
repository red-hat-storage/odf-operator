define DEPLOYMENT_ENV_PATCH
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  template:
    spec:
      containers:
      - name: manager
        env:
        - name: PKGS_CONFIG_MAP_NAME
          value: odf-operator-pkgs-config-$(VERSION)
endef
export DEPLOYMENT_ENV_PATCH


define PKGS_CONFIGMAP_YAML
apiVersion: v1
kind: ConfigMap
metadata:
  name: pkgs-config-$(VERSION)
data:
  OCS_OPERATOR: |
    package: $(OCS_OPERATOR_PKG_NAME)
    version: $(OCS_OPERATOR_PKG_VERSION)

  ROOK_CEPH: |
    package: $(ROOK_CEPH_PKG_NAME)
    version: $(ROOK_CEPH_PKG_VERSION)

  NOOBAA: |
    package: $(NOOBAA_PKG_NAME)
    version: $(NOOBAA_PKG_VERSION)

  OCS_CLIENT: |
    package: $(OCS_CLIENT_PKG_NAME)
    version: $(OCS_CLIENT_PKG_VERSION)

  CEPHCSI: |
    package: $(CEPHCSI_PKG_NAME)
    version: $(CEPHCSI_PKG_VERSION)

  CSIADDONS: |
    package: $(CSIADDONS_PKG_NAME)
    version: $(CSIADDONS_PKG_VERSION)

  ODF_SNAPSHOT: |
    package: $(ODF_SNAPSHOT_PKG_NAME)
    version: $(ODF_SNAPSHOT_PKG_VERSION)

  PROMETHEUS: |
    package: $(PROMETHEUS_PKG_NAME)
    version: $(PROMETHEUS_PKG_VERSION)

  OCS_TLS: |
    package: $(OCS_TLS_PKG_NAME)
    version: $(OCS_TLS_PKG_VERSION)

  RECIPE: |
    package: $(RECIPE_PKG_NAME)
    version: $(RECIPE_PKG_VERSION)

  IBM_ODF: |
    package: $(IBM_ODF_PKG_NAME)
    version: $(IBM_ODF_PKG_VERSION)

  IBM_CSI: |
    package: $(IBM_CSI_PKG_NAME)
    version: $(IBM_CSI_PKG_VERSION)

  CNSA: |
    package: $(CNSA_PKG_NAME)
    version: $(CNSA_PKG_VERSION)
    namespace: ibm-spectrum-scale
endef
export PKGS_CONFIGMAP_YAML
