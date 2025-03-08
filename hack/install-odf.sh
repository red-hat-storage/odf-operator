#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

INSTALL_NAMESPACE=openshift-storage
OPERATOR_SDK=${OPERATOR_SDK:-$1}
BUNDLE_IMG=${BUNDLE_IMG:-$2}
ODF_DEPS_BUNDLE_IMG=${ODF_DEPS_BUNDLE_IMG:-$3}
CATALOG_DEPS_IMG=${CATALOG_DEPS_IMG:-$4}
CSV_NAMES=${CSV_NAMES:-$5}
CI=${CI:-${6:-false}}

# If running in OpenShift CI, create an ICSP for the ODF dependencies bundle image.
# This ensures the cluster uses the CI-built image instead of pulling from an external registry.
# The image is defined in `catalog/odf-dependencies.yaml` and cannot be substituted during the deps-catalog build.

# The fix is needed because CI pulls the `latest` tag in all branches, if `latest` points to 4.19, tests pass on 4.19 but fails on others branches.
# This affects ODF 4.18 and above, as the ODF dependencies bundle was introduced in 4.18.

if [ "$CI" == true ]; then
    echo "
    apiVersion: operator.openshift.io/v1alpha1
    kind: ImageContentSourcePolicy
    metadata:
      name: odf-ci-images
    spec:
      repositoryDigestMirrors:
      - mirrors:
        - $ODF_DEPS_BUNDLE_IMG
        source: quay.io/ocs-dev/odf-dependencies-bundle:latest
    " | oc apply -f -
fi

NAMESPACE=$(oc get ns "$INSTALL_NAMESPACE" -o jsonpath="{.metadata.name}" 2>/dev/null || true)
if [[ -n "$NAMESPACE" ]]; then
    echo "Namespace \"$NAMESPACE\" exists"
else
    echo "Namespace \"$INSTALL_NAMESPACE\" does not exist: creating it"
    oc create ns "$INSTALL_NAMESPACE"
fi

"$OPERATOR_SDK" run bundle "$BUNDLE_IMG" --timeout=10m --security-context-config restricted -n "$INSTALL_NAMESPACE" --index-image "$CATALOG_DEPS_IMG"

sleep 30m

# Check for the presence of the CSVs in the cluster for up to 5 minutes,
# Since 'oc wait' exits immediately if the resource is not found.
for i in {1..30}; do
    if oc get -n "$INSTALL_NAMESPACE" csv $CSV_NAMES &> /dev/null; then
        break
    fi
    sleep 10
done

oc wait --timeout=5m --for jsonpath='{.status.phase}'=Succeeded -n "$INSTALL_NAMESPACE" csv $CSV_NAMES || {

    echo "CSV $CSV_NAMES did not succeed, describing CSV"
    oc get csv -n "$INSTALL_NAMESPACE"
    oc get pods -n "$INSTALL_NAMESPACE"
    oc describe csv -n "$INSTALL_NAMESPACE"
    oc describe pods -n "$INSTALL_NAMESPACE"
    exit 1
}

oc wait --timeout=5m --for condition=Available -n "$INSTALL_NAMESPACE" deployment \
    ceph-csi-controller-manager \
    csi-addons-controller-manager \
    noobaa-operator \
    ocs-client-operator-console \
    ocs-client-operator-controller-manager \
    ocs-operator \
    odf-console \
    odf-operator-controller-manager \
    prometheus-operator \
    rook-ceph-operator \
    ux-backend-server
