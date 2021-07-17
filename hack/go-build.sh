#!/bin/bash

CGO_ENABLED=${CGO_ENABLED:-0}
GOOS=${GOOS:-linux}
GOARCH=${GOARCH:-amd64}
GO111MODULE=${GO111MODULE:-on}

set -x

go build -a -o ${GOBIN:-bin}/manager main.go
