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


define CATALOG_INDEX_YAML
---
defaultChannel: $(DEFAULT_CHANNEL)
name: $(IMAGE_NAME)
schema: olm.package
---
schema: olm.channel
package: $(IMAGE_NAME)
name: $(DEFAULT_CHANNEL)
entries:
  - name: $(IMAGE_NAME).v$(VERSION)

---
defaultChannel: $(OCS_OPERATOR_PKG_CHANNEL)
name: $(OCS_OPERATOR_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(OCS_OPERATOR_PKG_NAME)
name: $(OCS_OPERATOR_PKG_CHANNEL)
entries:
  - name: $(OCS_OPERATOR_PKG_NAME).v$(OCS_OPERATOR_PKG_VERSION)

---
defaultChannel: $(ROOK_CEPH_PKG_CHANNEL)
name: $(ROOK_CEPH_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(ROOK_CEPH_PKG_NAME)
name: $(ROOK_CEPH_PKG_CHANNEL)
entries:
  - name: $(ROOK_CEPH_PKG_NAME).v$(ROOK_CEPH_PKG_VERSION)

---
defaultChannel: $(NOOBAA_PKG_CHANNEL)
name: $(NOOBAA_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(NOOBAA_PKG_NAME)
name: $(NOOBAA_PKG_CHANNEL)
entries:
  - name: $(NOOBAA_PKG_NAME).v$(NOOBAA_PKG_VERSION)

---
defaultChannel: $(OCS_CLIENT_PKG_CHANNEL)
name: $(OCS_CLIENT_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(OCS_CLIENT_PKG_NAME)
name: $(OCS_CLIENT_PKG_CHANNEL)
entries:
  - name: $(OCS_CLIENT_PKG_NAME).v$(OCS_CLIENT_PKG_VERSION)

---
defaultChannel: $(CEPHCSI_PKG_CHANNEL)
name: $(CEPHCSI_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(CEPHCSI_PKG_NAME)
name: $(CEPHCSI_PKG_CHANNEL)
entries:
  - name: $(CEPHCSI_PKG_NAME).v$(CEPHCSI_PKG_VERSION)

---
defaultChannel: $(CSIADDONS_PKG_CHANNEL)
name: $(CSIADDONS_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(CSIADDONS_PKG_NAME)
name: $(CSIADDONS_PKG_CHANNEL)
entries:
  - name: $(CSIADDONS_PKG_NAME).v$(CSIADDONS_PKG_VERSION)

---
defaultChannel: $(ODF_SNAPSHOT_PKG_CHANNEL)
name: $(ODF_SNAPSHOT_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(ODF_SNAPSHOT_PKG_NAME)
name: $(ODF_SNAPSHOT_PKG_CHANNEL)
entries:
  - name: $(ODF_SNAPSHOT_PKG_NAME).v$(ODF_SNAPSHOT_PKG_VERSION)

---
defaultChannel: $(PROMETHEUS_PKG_CHANNEL)
name: $(PROMETHEUS_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(PROMETHEUS_PKG_NAME)
name: $(PROMETHEUS_PKG_CHANNEL)
entries:
  - name: $(PROMETHEUS_PKG_NAME).v$(PROMETHEUS_PKG_VERSION)

---
defaultChannel: $(OCS_TLS_PKG_CHANNEL)
name: $(OCS_TLS_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(OCS_TLS_PKG_NAME)
name: $(OCS_TLS_PKG_CHANNEL)
entries:
  - name: $(OCS_TLS_PKG_NAME).v$(OCS_TLS_PKG_VERSION)

---
defaultChannel: $(RECIPE_PKG_CHANNEL)
name: $(RECIPE_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(RECIPE_PKG_NAME)
name: $(RECIPE_PKG_CHANNEL)
entries:
  - name: $(RECIPE_PKG_NAME).v$(RECIPE_PKG_VERSION)

---
defaultChannel: $(IBM_ODF_PKG_CHANNEL)
name: $(IBM_ODF_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(IBM_ODF_PKG_NAME)
name: $(IBM_ODF_PKG_CHANNEL)
entries:
  - name: $(IBM_ODF_PKG_NAME).v$(IBM_ODF_PKG_VERSION)

---
defaultChannel: $(IBM_CSI_PKG_CHANNEL)
name: $(IBM_CSI_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(IBM_CSI_PKG_NAME)
name: $(IBM_CSI_PKG_CHANNEL)
entries:
  - name: $(IBM_CSI_PKG_NAME).v$(IBM_CSI_PKG_VERSION)

---
defaultChannel: $(CNSA_PKG_CHANNEL)
name: $(CNSA_PKG_NAME)
schema: olm.package
---
schema: olm.channel
package: $(CNSA_PKG_NAME)
name: $(CNSA_PKG_CHANNEL)
entries:
  - name: $(CNSA_PKG_NAME).v$(CNSA_PKG_VERSION)
endef
export CATALOG_INDEX_YAML
