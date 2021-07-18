# go-get-tool will 'go get' any package $2 and install it to $1.
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

##@ Tools

CONTROLLER_GEN = $(BIN_DIR)/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(BIN_DIR)/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

OPERATOR_SDK = $(BIN_DIR)/operator-sdk
operator-sdk: ## Download operator-sdk locally if necessary.
	@./hack/get-tool.sh $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/v1.9.0/operator-sdk_$(GOOS)_$(GOARCH)

.PHONY: opm
OPM = $(BIN_DIR)/opm
opm: ## Download opm locally if necessary.
	@./hack/get-tool.sh $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$(GOOS)-$(GOARCH)-opm
