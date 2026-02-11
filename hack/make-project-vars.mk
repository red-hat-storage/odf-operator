PROJECT_DIR := $(PWD)
BIN_DIR := $(PROJECT_DIR)/bin
ENVTEST_K8S_VERSION := 1.27.1

GOBIN ?= $(BIN_DIR)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPROXY ?= https://proxy.golang.org/

GO_LINT_IMG_LOCATION ?= golangci/golangci-lint
GO_LINT_IMG_TAG ?= v1.64.5
GO_LINT_IMG ?= $(GO_LINT_IMG_LOCATION):$(GO_LINT_IMG_TAG)
