#!/bin/bash

set -x

cd e2e/odf && ${GINKGO} build && ./odf.test \
    --odf-catalog-image=${CATALOG_IMG} \
    --odf-subscription-channel=${CHANNELS} \
    --odf-operator-install=${ODF_OPERATOR_INSTALL} \
    --odf-operator-uninstall=${ODF_OPERATOR_UNINSTALL} \
    --odf-cluster-service-version=odf-operator.v${VERSION} \
    --odf-deps-cluster-service-version=odf-dependencies.v${VERSION} \
    --ocs-cluster-service-version=${OCS_SUBSCRIPTION_STARTINGCSV} \
    --ocs-client-cluster-service-version=${OCS_CLIENT_SUBSCRIPTION_STARTINGCSV} \
    --nooba-cluster-service-version=${NOOBAA_SUBSCRIPTION_STARTINGCSV} \
    --csiaddons-cluster-service-version=${CSIADDONS_SUBSCRIPTION_STARTINGCSV} \
    --cephcsi-cluster-service-version=${CEPHCSI_SUBSCRIPTION_STARTINGCSV} \
    --rook-cluster-service-version=${ROOK_SUBSCRIPTION_STARTINGCSV} \
    --prometheus-cluster-service-version=${PROMETHEUS_SUBSCRIPTION_STARTINGCSV} \
    --recipe-cluster-service-version=${RECIPE_SUBSCRIPTION_STARTINGCSV}
