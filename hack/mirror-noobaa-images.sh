#!/usr/bin/env bash
set -eu
set -o pipefail
shopt -s inherit_errexit
set -x

# Script will mirror the latest NooBaa upstream master images to quay.io/ocs-dev with a 'master' tag
# that can be used by CI.

# Requirements:
#   - cli: jq docker
#   - must be logged into quay to push to quay.io/ocs-dev


function latest_master_tag () {
    local repo="$1"

    URL=https://registry.hub.docker.com/v2/repositories/${repo}/tags/
    # noobaa images are tagged as "master-<date>" (e.g., master-20210824)
    tags="$(curl -s -S "$URL" | jq --raw-output '.results[].name')"
    master_tags="$(echo "${tags}" | grep --extended-regexp "^master-[[:digit:]]{8}$")"
    latest_tag="$(echo "${master_tags}" | sort --numeric-sort --reverse | head -n1 )"
    echo "${latest_tag}"
}

for image in noobaa-core noobaa-operator; do
    repo="noobaa/${image}"
    latest="$(latest_master_tag "${repo}")"
    latest_image="${repo}:${latest}"

    docker pull "${latest_image}"

    quay_latest="quay.io/ocs-dev/${image}:${latest}"
    quay_master="quay.io/ocs-dev/${image}:master"
    docker tag "${latest_image}" "${quay_latest}"
    docker tag "${latest_image}" "${quay_master}"

    docker push "${quay_latest}"
    docker push "${quay_master}"
done
