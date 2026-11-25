#!/bin/bash
set -euxo pipefail
shopt -s inherit_errexit 2>/dev/null || true

if [[ -z "${KUBECONFIG:-}" ]]; then
    echo "KUBECONFIG not set, skipping" >&2
    exit 2
fi

csvName=$(oc get csv --all-namespaces --no-headers 2>/dev/null | grep -E "odf-operator|ocs-operator" | head -1 | awk '{print $2}' || true)
if [[ -z "${csvName}" ]]; then
    echo "WARN: ODF / OCS operator CSV not found — ODF may not be deployed" >&2
    exit 2
fi

csvPhase=$(oc get csv --all-namespaces -o jsonpath="{.items[?(@.metadata.name==\"${csvName}\")].status.phase}" 2>/dev/null | awk '{print $1}')
if [[ "${csvPhase}" != "Succeeded" ]]; then
    echo "FAIL: ODF operator CSV is not Succeeded (phase: ${csvPhase})" >&2
    exit 1
fi

echo "PASS: ODF operator is installed (${csvName})"
exit 0
