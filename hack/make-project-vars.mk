PROJECT_DIR := $(PWD)
BIN_DIR := $(PROJECT_DIR)/bin
ENVTEST_ASSETS_DIR := $(PROJECT_DIR)/testbin

GOBIN ?= $(BIN_DIR)
GOOS ?= linux
GOARCH ?= amd64
GOPROXY ?= https://proxy.golang.org/

GO_LINT_IMG_LOCATION ?= golangci/golangci-lint
GO_LINT_IMG_TAG ?= v1.49.0
GO_LINT_IMG ?= $(GO_LINT_IMG_LOCATION):$(GO_LINT_IMG_TAG)
