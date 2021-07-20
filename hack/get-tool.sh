#!/bin/bash

BIN="${1}"
URL="${2}"

set -e

if [ ! -f "${BIN}" ]; then
  echo "Downloading ${URL}"
  mkdir -p "$(dirname "${BIN}")"
  curl -sSLo "${BIN}" "${URL}"
  chmod +x "${BIN}"
fi
