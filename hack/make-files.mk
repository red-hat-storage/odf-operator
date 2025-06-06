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


# In external storageCluster there won't be any storageClient but CSI is managed by client op hence we need to
# scale up client op based on cephCluster instead of storageClient CR
define CONFIGMAP_YAML
apiVersion: v1
kind: ConfigMap
metadata:
  name: pkgs-config-$(VERSION)
data:
  CEPHCSI: |
    channel: $(CEPHCSI_SUBSCRIPTION_CHANNEL)
    csv: $(CEPHCSI_SUBSCRIPTION_CSVNAME)
    pkg: $(CEPHCSI_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
  CSIADDONS: |
    channel: $(CSIADDONS_SUBSCRIPTION_CHANNEL)
    csv: $(CSIADDONS_SUBSCRIPTION_CSVNAME)
    pkg: $(CSIADDONS_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
  SNAPSHOT_CONTROLLER: |
    channel: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_CHANNEL)
    csv: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_CSVNAME)
    pkg: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
  IBM_ODF: |
    channel: $(IBM_ODF_SUBSCRIPTION_CHANNEL)
    csv: $(IBM_ODF_SUBSCRIPTION_CSVNAME)
    pkg: $(IBM_ODF_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - flashsystemclusters.odf.ibm.com
  NOOBAA: |
    channel: $(NOOBAA_SUBSCRIPTION_CHANNEL)
    csv: $(NOOBAA_SUBSCRIPTION_CSVNAME)
    pkg: $(NOOBAA_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - noobaas.noobaa.io
  OCS_CLIENT: |
    channel: $(OCS_CLIENT_SUBSCRIPTION_CHANNEL)
    csv: $(OCS_CLIENT_SUBSCRIPTION_CSVNAME)
    pkg: $(OCS_CLIENT_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
  OCS: |
    channel: $(OCS_SUBSCRIPTION_CHANNEL)
    csv: $(OCS_SUBSCRIPTION_CSVNAME)
    pkg: $(OCS_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - storageclusters.ocs.openshift.io
  ODF_DEPS: |
    channel: $(ODF_DEPS_SUBSCRIPTION_CHANNEL)
    csv: $(ODF_DEPS_SUBSCRIPTION_CSVNAME)
    pkg: $(ODF_DEPS_SUBSCRIPTION_PACKAGE)
  PROMETHEUS: |
    channel: $(PROMETHEUS_SUBSCRIPTION_CHANNEL)
    csv: $(PROMETHEUS_SUBSCRIPTION_CSVNAME)
    pkg: $(PROMETHEUS_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - alertmanagers.monitoring.coreos.com
      - prometheuses.monitoring.coreos.com
  RECIPE: |
    channel: $(RECIPE_SUBSCRIPTION_CHANNEL)
    csv: $(RECIPE_SUBSCRIPTION_CSVNAME)
    pkg: $(RECIPE_SUBSCRIPTION_PACKAGE)
  ROOK: |
    channel: $(ROOK_SUBSCRIPTION_CHANNEL)
    csv: $(ROOK_SUBSCRIPTION_CSVNAME)
    pkg: $(ROOK_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
endef
export CONFIGMAP_YAML


define DEPENDENCIES_YAML
dependencies:
- type: olm.package
  value:
    packageName: $(OCS_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(OCS_BUNDLE_VERSION))"
- type: olm.package
  value:
    packageName: $(ROOK_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(ROOK_BUNDLE_VERSION))"
- type: olm.package
  value:
    packageName: $(OCS_CLIENT_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(OCS_CLIENT_BUNDLE_VERSION))"
- type: olm.package
  value:
    packageName: $(NOOBAA_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(NOOBAA_BUNDLE_VERSION))"
- type: olm.package
  value:
    packageName: $(CSIADDONS_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(CSIADDONS_BUNDLE_VERSION))"
- type: olm.package
  value:
    packageName: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(ODF_SNAPSHOT_CONTROLLER_BUNDLE_VERSION))"
- type: olm.package
  value:
    packageName: $(CEPHCSI_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(CEPHCSI_BUNDLE_VERSION))"
- type: olm.package
  value:
    packageName: $(PROMETHEUS_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(PROMETHEUS_BUNDLE_VERSION))"
- type: olm.package
  value:
    packageName: $(RECIPE_SUBSCRIPTION_PACKAGE)
    version: "$(subst v,,$(RECIPE_BUNDLE_VERSION))"
endef
export DEPENDENCIES_YAML


define INDEX_YAML
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
defaultChannel: $(ODF_DEPS_SUBSCRIPTION_CHANNEL)
name: $(ODF_DEPS_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(ODF_DEPS_SUBSCRIPTION_PACKAGE)
name: $(ODF_DEPS_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(ODF_DEPS_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(OCS_SUBSCRIPTION_CHANNEL)
name: $(OCS_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(OCS_SUBSCRIPTION_PACKAGE)
name: $(OCS_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(OCS_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(OCS_CLIENT_SUBSCRIPTION_CHANNEL)
name: $(OCS_CLIENT_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(OCS_CLIENT_SUBSCRIPTION_PACKAGE)
name: $(OCS_CLIENT_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(OCS_CLIENT_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(NOOBAA_SUBSCRIPTION_CHANNEL)
name: $(NOOBAA_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(NOOBAA_SUBSCRIPTION_PACKAGE)
name: $(NOOBAA_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(NOOBAA_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(CSIADDONS_SUBSCRIPTION_CHANNEL)
name: $(CSIADDONS_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(CSIADDONS_SUBSCRIPTION_PACKAGE)
name: $(CSIADDONS_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(CSIADDONS_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_CHANNEL)
name: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_PACKAGE)
name: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(ODF_SNAPSHOT_CONTROLLER_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(CEPHCSI_SUBSCRIPTION_CHANNEL)
name: $(CEPHCSI_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(CEPHCSI_SUBSCRIPTION_PACKAGE)
name: $(CEPHCSI_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(CEPHCSI_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(ROOK_SUBSCRIPTION_CHANNEL)
name: $(ROOK_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(ROOK_SUBSCRIPTION_PACKAGE)
name: $(ROOK_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(ROOK_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(PROMETHEUS_SUBSCRIPTION_CHANNEL)
name: $(PROMETHEUS_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(PROMETHEUS_SUBSCRIPTION_PACKAGE)
name: $(PROMETHEUS_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(PROMETHEUS_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(RECIPE_SUBSCRIPTION_CHANNEL)
name: $(RECIPE_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(RECIPE_SUBSCRIPTION_PACKAGE)
name: $(RECIPE_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(RECIPE_SUBSCRIPTION_CSVNAME)

---
defaultChannel: $(IBM_ODF_SUBSCRIPTION_CHANNEL)
name: $(IBM_ODF_SUBSCRIPTION_PACKAGE)
schema: olm.package
---
schema: olm.channel
package: $(IBM_ODF_SUBSCRIPTION_PACKAGE)
name: $(IBM_ODF_SUBSCRIPTION_CHANNEL)
entries:
  - name: $(IBM_ODF_SUBSCRIPTION_CSVNAME)
endef
export INDEX_YAML
