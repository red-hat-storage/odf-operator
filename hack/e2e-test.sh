#!/bin/bash

set -x

cd e2e/odf && ${GINKGO} build && ./odf.test \
    --odf-catalog-image=${CATALOG_IMG} \
    --odf-subscription-channel=${CHANNELS} \
    --odf-operator-install=${ODF_OPERATOR_INSTALL} \
    --odf-operator-uninstall=${ODF_OPERATOR_UNINSTALL} \
    --odf-cluster-service-version=odf-operator.v${VERSION} \
    --csv-names="${CSV_NAMES}"
