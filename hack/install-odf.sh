#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

INSTALL_NAMESPACE=openshift-storage
OPERATOR_SDK=${OPERATOR_SDK:-$1}
BUNDLE_IMG=${BUNDLE_IMG:-$2}
CATALOG_DEPS_IMG=${CATALOG_DEPS_IMG:-$3}
CSV_NAMES=${CSV_NAMES:-$4}

NAMESPACE=$(oc get ns "$INSTALL_NAMESPACE" -o jsonpath="{.metadata.name}" 2>/dev/null || true)
if [[ -n "$NAMESPACE" ]]; then
    echo "Namespace \"$NAMESPACE\" exists"
else
    echo "Namespace \"$INSTALL_NAMESPACE\" does not exist: creating it"
    oc create ns "$INSTALL_NAMESPACE"
fi

"$OPERATOR_SDK" run bundle "$BUNDLE_IMG" --timeout=10m --security-context-config restricted -n "$INSTALL_NAMESPACE" --index-image "$CATALOG_DEPS_IMG"

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
