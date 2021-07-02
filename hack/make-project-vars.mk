PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
