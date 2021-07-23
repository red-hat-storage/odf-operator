#!/bin/bash

export CGO_ENABLED=${CGO_ENABLED:-0}
export GOOS=${GOOS:-linux}
export GOARCH=${GOARCH:-amd64}
export GO111MODULE=${GO111MODULE:-on}

set -x

go build -a -o ${GOBIN:-bin}/manager main.go
