#!/usr/bin/env bash

set -e
set -x

pushd services/devicefinder
podman build -f Dockerfile -t "${DEVICEFINDER_IMAGE}" ../.. \
    --build-arg="LDFLAGS=${LDFLAGS}" --platform="linux/amd64"
popd
