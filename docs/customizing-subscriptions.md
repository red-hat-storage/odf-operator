## Customize odf-operator to create customized subscriptions

odf-operator manages 3 subscriptions `ocs-operator`, `noobaa-operator` and
`ibm-storage-odf-operator` which can be customized proactively during build
or via editing configmap later.

### Customize odf-operator during build

We can customize odf-operator to create or update the subscriptions via
exporting env variables while running make commands. For example we can create
our own catalogsource with ocs bundle in the OCP cluster and change the
`OCS_SUBSCRIPTION_CATALOGSOURCE` and `OCS_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE`
to use our own catalogsource.

Build a bundle image with the customized values via:
```
OCS_SUBSCRIPTION_CATALOGSOURCE=redhat-operators \
OCS_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE=openshift-marketplace \
make bundle-build
```

### Customize odf-operator via editing configmap

We can also customize odf-operator to create or update the subscriptions via
values in a ConfigMap. For example we can create our own catalogsource with
ocs bundle in the OCP cluster and change the `OCS_SUBSCRIPTION_CATALOGSOURCE`
and `OCS_SUBSCRIPTION_CATALOGSOURCE_NAMESPACE` to use our own catalogsource.

Edit the configmap to change the default values for subscriptions via:
```
oc edit configmap odf-operator-manager-config
```

**Note:** You can see all variables and their default values [here](
../bundle/manifests/odf-operator-manager-config_v1_configmap.yaml#L3-L22)

ConfigMaps consumed as environment variables are not updated automatically and
require a pod restart. Restart the operator `odf-operator-controller-manager`
via deleting it to consume the new values.
