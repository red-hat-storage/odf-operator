#!/bin/bash

SCRIPT_DIR="$(dirname "$(realpath "$0")")"

source "${SCRIPT_DIR}/go-test-setup.sh"

set -x

go test -coverprofile cover.out `go list ./... | grep -v "e2e"`
