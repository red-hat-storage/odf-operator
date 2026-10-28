#!/bin/bash
set -euxo pipefail
shopt -s inherit_errexit 2>/dev/null || true

if [[ -z "${KUBECONFIG:-}" ]]; then
    echo "KUBECONFIG not set, skipping" >&2
    exit 2
fi

if ! oc get crd storageclusters.ocs.openshift.io &>/dev/null; then
    echo "WARN: CRD storageclusters.ocs.openshift.io is not installed — ODF may not be deployed" >&2
    exit 2
fi

echo "PASS: ODF StorageCluster CRD is installed"
exit 0
