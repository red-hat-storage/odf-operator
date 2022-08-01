#!/bin/bash

ENVTEST_ASSETS_DIR="${ENVTEST_ASSETS_DIR:-testbin}"
SKIP_FETCH_TOOLS="${SKIP_FETCH_TOOLS:-}"

mkdir -p "${ENVTEST_ASSETS_DIR}"

pushd "${ENVTEST_ASSETS_DIR}" > /dev/null


if [ ! -f setup-envtest.sh ]; then
	curl -sSLo setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
fi

source setup-envtest.sh

fetch_envtest_tools "$(pwd)"
setup_envtest_env "$(pwd)"

popd > /dev/null
