include hack/make-project-vars.mk
include hack/make-tools.mk
include hack/make-bundle-vars.mk
include hack/make-files.mk


# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.DEFAULT_GOAL := help
.EXPORT_ALL_VARIABLES:

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@./hack/make-help.sh $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen update-mgr-config ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

lint: ## Run golangci-lint against code.
	docker run --rm -v $(PROJECT_DIR):/app:Z -w /app $(GO_LINT_IMG) golangci-lint run -E gosec --timeout=6m .

godeps-update: ## Run go mod tidy and go mod vendor.
	go mod tidy && go mod vendor

test-setup: generate fmt vet godeps-update ## Run setup targets for tests

go-test: ## Run go test against code.
	./hack/go-test.sh

test: test-setup go-test ## Run go unit tests.

ODF_OPERATOR_INSTALL ?= false
ODF_OPERATOR_UNINSTALL ?= false
# PKGS_CONFIG_MAP_NAME is used by the ODF operator to read subscription-related configuration.
# It must be set before the test binary runs, so that dependent packages like deploymanager
# can read it during their execution.
e2e-test: export PKGS_CONFIG_MAP_NAME=odf-operator-pkgs-config-${VERSION}
e2e-test: ginkgo ## Run end to end functional tests.
	@echo "build and run e2e tests"
	cd e2e/odf && ${GINKGO} build && ./odf.test --ginkgo.v \
		--odf-operator-install=${ODF_OPERATOR_INSTALL} \
		--odf-operator-uninstall=${ODF_OPERATOR_UNINSTALL} \
		--odf-catalog-image=${CATALOG_IMG} \
		--odf-subscription-channel=${CHANNELS} \
		--odf-cluster-service-version=odf-operator.v${VERSION} \
		--csv-names="${CSV_NAMES}"

update-mgr-config: ## Feed env variables to the manager configmap
	@echo "$$DEPLOYMENT_ENV_PATCH" > config/manager/deployment-env-patch.yaml
	@echo "$$CONFIGMAP_YAML" > config/manager/configmap.yaml

# ------------------------------------------------------------------------------
# This target prints additional FDF dependencies. This will be used in the DS build process.
#
# IMPORTANT:
# - Printing should only be enabled once the Fusion is available for the target OCP/ODF release.
#
# HOW TO CONTROL PRINTING:
# - To enable printing, set the IS_FUSION_PRESENT=true.
# - To disable printing, set the IS_FUSION_PRESENT=false.
# ------------------------------------------------------------------------------
IS_FUSION_PRESENT=true
gen-additional-fdf-dependencies:
	@[ "$(IS_FUSION_PRESENT)" = "true" ] && echo "$$ADDITIONAL_FDF_DEPENDENCIES_YAML" || true

##@ Build

build: generate fmt vet go-build ## Build manager binary.

go-build: ## Run go build against code.
	@GOBIN=${GOBIN} ./hack/go-build.sh

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

docker-build: godeps-update test-setup go-test ## Build docker image with the manager.
	docker build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

install-odf: operator-sdk ## install odf using the hack/install-odf.sh script
	hack/install-odf.sh $(OPERATOR_SDK) $(BUNDLE_IMG) $(ODF_DEPS_CATALOG_IMG) "$(CSV_NAMES)"

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	cd config/console && $(KUSTOMIZE) edit set image odf-console=$(ODF_CONSOLE_IMG)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

deploy-with-olm: kustomize ## Deploy controller to the K8s cluster via OLM
	cd config/install && $(KUSTOMIZE) edit set image catalog-img=${CATALOG_IMG}
	cd config/install/odf-resources && $(KUSTOMIZE) edit set namespace $(OPERATOR_NAMESPACE)
	$(KUSTOMIZE) build config/install | kubectl create -f -

undeploy-with-olm: ## Undeploy controller from the K8s cluster
	$(KUSTOMIZE) build config/install | kubectl delete -f -

checkout-bundle-timestamp: ## Ignore (git checkout) changes if there are only timestamp changes in the bundle
	(git diff --quiet --ignore-matching-lines createdAt bundle/$(BUNDLE_DIR) && git checkout --quiet bundle/$(BUNDLE_DIR)) || true

bundle-dependencies: ## Generate dependencies bundle manifests and metadata, then validate generated files.
	@test -n "$$DEPS_YAML" || (echo "ERROR: DEPS_YAML is not set."; exit 1)
	@test -n "$(PKG_NAME)" || (echo "ERROR: PKG_NAME is not set."; exit 1)
	@test -n "$(DOCKERFILE_NAME)" || (echo "ERROR: DOCKERFILE_NAME is not set."; exit 1)

	@echo "$$DEPS_YAML" > bundle/$(PKG_NAME)/metadata/dependencies.yaml
	cd config/bundles/$(PKG_NAME) && $(KUSTOMIZE) edit add annotation --force \
		'olm.skipRange':"$(SKIP_RANGE)" \
		'olm.properties':'[{"type": "olm.maxOpenShiftVersion", "value": "$(MAX_OCP_VERSION)"}]' && \
		$(KUSTOMIZE) edit add patch --name $(PKG_NAME).v0.0.0 --kind ClusterServiceVersion \
		--patch '[{"op": "replace", "path": "/spec/replaces", "value": "$(REPLACES)"}]'
	$(KUSTOMIZE) build config/bundles/$(PKG_NAME) | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS) \
		--output-dir bundle/$(PKG_NAME) --package $(PKG_NAME)
	$(OPERATOR_SDK) bundle validate bundle/$(PKG_NAME)
	@$(MAKE) --no-print-directory checkout-bundle-timestamp BUNDLE_DIR=$(PKG_NAME)
	@mv bundle.Dockerfile $(DOCKERFILE_NAME)

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	# cnsa dependencies bundle
	$(MAKE) bundle-dependencies \
		DEPS_YAML="$$CNSA_DEPENDENCIES_YAML" \
		PKG_NAME="cnsa-dependencies" \
		DOCKERFILE_NAME="bundle.cnsa.deps.Dockerfile"

	# odf dependencies bundle
	$(MAKE) bundle-dependencies \
		DEPS_YAML="$$ODF_DEPENDENCIES_YAML" \
		PKG_NAME="odf-dependencies" \
		DOCKERFILE_NAME="bundle.odf.deps.Dockerfile"

	# Main odf-operator bundle
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	cd config/console && $(KUSTOMIZE) edit set image odf-console=$(ODF_CONSOLE_IMG)
	cd config/manifests/bases && $(KUSTOMIZE) edit add annotation --force \
		'olm.skipRange':"$(SKIP_RANGE)" \
		'olm.properties':'[{"type": "olm.maxOpenShiftVersion", "value": "$(MAX_OCP_VERSION)"}]' && \
		$(KUSTOMIZE) edit add patch --name odf-operator.v0.0.0 --kind ClusterServiceVersion \
		--patch '[{"op": "replace", "path": "/spec/replaces", "value": "$(REPLACES)"}]'
	rm -rf bundle/odf-operator/manifests
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS) \
		--output-dir bundle/odf-operator
	$(OPERATOR_SDK) bundle validate bundle/odf-operator
	@$(MAKE) --no-print-directory checkout-bundle-timestamp BUNDLE_DIR=odf-operator

.PHONY: bundle-build
bundle-build: bundle ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .
	docker build -f bundle.odf.deps.Dockerfile -t $(ODF_DEPS_BUNDLE_IMG) .
	docker build -f bundle.cnsa.deps.Dockerfile -t $(CNSA_DEPS_BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)
	$(MAKE) docker-push IMG=$(ODF_DEPS_BUNDLE_IMG)
	$(MAKE) docker-push IMG=$(CNSA_DEPS_BUNDLE_IMG)

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog
catalog: opm ## Generate catalog manifests and then validate generated files.
	@echo "$$INDEX_YAML" > catalog/index.yaml
	$(OPM) render --output=yaml $(BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/odf.yaml
	$(OPM) render --output=yaml $(ODF_DEPS_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/odf-dependencies.yaml
	$(OPM) render --output=yaml $(CNSA_DEPS_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/cnsa-dependencies.yaml
	$(OPM) render --output=yaml $(OCS_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/ocs.yaml
	$(OPM) render --output=yaml $(OCS_CLIENT_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/ocs-client.yaml
	$(OPM) render --output=yaml $(IBM_ODF_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/ibm.yaml
	$(OPM) render --output=yaml $(IBM_CSI_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/ibm-csi.yaml
	$(OPM) render --output=yaml $(CNSA_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/cnsa.yaml
	$(OPM) render --output=yaml $(NOOBAA_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/noobaa.yaml
	$(OPM) render --output=yaml $(CSIADDONS_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/csiaddons.yaml
	$(OPM) render --output=yaml $(ODF_SNAPSHOT_CONTROLLER_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/snapshotter.yaml
	$(OPM) render --output=yaml $(CEPHCSI_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/cephcsi.yaml
	$(OPM) render --output=yaml $(ROOK_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/rook.yaml
	$(OPM) render --output=yaml $(PROMETHEUS_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/prometheus.yaml
	$(OPM) render --output=yaml $(RECIPE_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/recipe.yaml
	$(OPM) validate catalog

.PHONY: catalog-build
catalog-build: catalog ## Build a catalog image.
	docker build -f catalog.Dockerfile -t $(CATALOG_IMG) .

.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

.PHONY: catalog-deps-build
catalog-deps-build: catalog ## Build a catalog-deps image.
	docker build -f catalog.deps.Dockerfile -t $(ODF_DEPS_CATALOG_IMG) .

.PHONY: catalog-deps-push
catalog-deps-push: ## Push a catalog-deps image.
	$(MAKE) docker-push IMG=$(ODF_DEPS_CATALOG_IMG)
