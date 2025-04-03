include hack/make-project-vars.mk
include hack/make-tools.mk
include hack/make-bundle-vars.mk


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

manifests: controller-gen update-mgr-env ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
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

ODF_OPERATOR_INSTALL ?= true
ODF_OPERATOR_UNINSTALL ?= true
e2e-test: ginkgo ## Run end to end functional tests.
	@echo "build and run e2e tests"
	./hack/e2e-test.sh

# In external storageCluster there won't be any storageClient but CSI is managed by client op hence we need to
# scale up client op based on cephCluster instead of storageClient CR
define MAPPING
apiVersion: v1
kind: ConfigMap
metadata:
  name: manager-config
data:
  CEPHCSI: |
    channel: $(CEPHCSI_SUBSCRIPTION_CHANNEL)
    csv: $(CEPHCSI_SUBSCRIPTION_STARTINGCSV)
    pkg: $(CEPHCSI_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
  CSIADDONS: |
    channel: $(CSIADDONS_SUBSCRIPTION_CHANNEL)
    csv: $(CSIADDONS_SUBSCRIPTION_STARTINGCSV)
    pkg: $(CSIADDONS_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
  IBM: |
    channel: $(IBM_SUBSCRIPTION_CHANNEL)
    csv: $(IBM_SUBSCRIPTION_STARTINGCSV)
    pkg: $(IBM_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - flashsystemclusters.odf.ibm.com
  NOOBAA: |
    channel: $(NOOBAA_SUBSCRIPTION_CHANNEL)
    csv: $(NOOBAA_SUBSCRIPTION_STARTINGCSV)
    pkg: $(NOOBAA_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - noobaas.noobaa.io
  OCS_CLIENT: |
    channel: $(OCS_CLIENT_SUBSCRIPTION_CHANNEL)
    csv: $(OCS_CLIENT_SUBSCRIPTION_STARTINGCSV)
    pkg: $(OCS_CLIENT_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
  OCS: |
    channel: $(OCS_SUBSCRIPTION_CHANNEL)
    csv: $(OCS_SUBSCRIPTION_STARTINGCSV)
    pkg: $(OCS_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - storageclusters.ocs.openshift.io
  ODF_DEPS: |
    channel: $(ODF_DEPS_SUBSCRIPTION_CHANNEL)
    csv: $(ODF_DEPS_SUBSCRIPTION_STARTINGCSV)
    pkg: $(ODF_DEPS_SUBSCRIPTION_PACKAGE)
  PROMETHEUS: |
    channel: $(PROMETHEUS_SUBSCRIPTION_CHANNEL)
    csv: $(PROMETHEUS_SUBSCRIPTION_STARTINGCSV)
    pkg: $(PROMETHEUS_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - alertmanagers.monitoring.coreos.com
      - prometheuses.monitoring.coreos.com
  RECIPE: |
    channel: $(RECIPE_SUBSCRIPTION_CHANNEL)
    csv: $(RECIPE_SUBSCRIPTION_STARTINGCSV)
    pkg: $(RECIPE_SUBSCRIPTION_PACKAGE)
  ROOK: |
    channel: $(ROOK_SUBSCRIPTION_CHANNEL)
    csv: $(ROOK_SUBSCRIPTION_STARTINGCSV)
    pkg: $(ROOK_SUBSCRIPTION_PACKAGE)
    scaleUpOnInstanceOf:
      - cephclusters.ceph.rook.io
endef
export MAPPING


update-mgr-env: ## Feed env variables to the manager configmap
	@echo "$$MAPPING" > config/manager/mapping.yaml

##@ Build

build: generate fmt vet go-build ## Build manager binary.

go-build: ## Run go build against code.
	@GOBIN=${GOBIN} ./hack/go-build.sh

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

docker-build: godeps-update test-setup ## Build docker image with the manager.
	docker build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

install-odf: operator-sdk ## install odf using the hack/install-odf.sh script
	hack/install-odf.sh $(OPERATOR_SDK) $(BUNDLE_IMG) $(ODF_DEPS_CATALOG_IMG) $(STARTING_CSVS)

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

# Make target to ignore (git checkout) changes if there are only timestamp changes in the bundle
checkout-bundle-timestamp:
	(git diff --quiet --ignore-matching-lines createdAt bundle/odf-operator && git checkout --quiet bundle/odf-operator) || true
	(git diff --quiet --ignore-matching-lines createdAt bundle/odf-dependencies && git checkout --quiet bundle/odf-dependencies) || true

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	# Dependencies bundle
	cd config/bundle && $(KUSTOMIZE) edit add annotation --force \
		'olm.skipRange':"$(SKIP_RANGE)" \
		'olm.properties':'[{"type": "olm.maxOpenShiftVersion", "value": "$(MAX_OCP_VERSION)"}]' && \
		$(KUSTOMIZE) edit add patch --name odf-dependencies.v0.0.0 --kind ClusterServiceVersion \
		--patch '[{"op": "replace", "path": "/spec/replaces", "value": "$(REPLACES)"}]'
	$(KUSTOMIZE) build config/bundle | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS) \
		--output-dir bundle/odf-dependencies --package odf-dependencies
	$(OPERATOR_SDK) bundle validate bundle/odf-dependencies
	@mv bundle.Dockerfile bundle.deps.Dockerfile

	# Main odf-operator bundle
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	cd config/console && $(KUSTOMIZE) edit set image odf-console=$(ODF_CONSOLE_IMG)
	cd config/manifests/bases && $(KUSTOMIZE) edit add annotation --force \
		'olm.skipRange':"$(SKIP_RANGE)" \
		'olm.properties':'[{"type": "olm.maxOpenShiftVersion", "value": "$(MAX_OCP_VERSION)"}]' && \
		$(KUSTOMIZE) edit add patch --name odf-operator.v0.0.0 --kind ClusterServiceVersion \
		--patch '[{"op": "replace", "path": "/spec/replaces", "value": "$(REPLACES)"}]'
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS) \
		--output-dir bundle/odf-operator
	$(OPERATOR_SDK) bundle validate bundle/odf-operator
	@$(MAKE) --no-print-directory checkout-bundle-timestamp

.PHONY: bundle-build
bundle-build: bundle ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .
	docker build -f bundle.deps.Dockerfile -t $(ODF_DEPS_BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)
	$(MAKE) docker-push IMG=$(ODF_DEPS_BUNDLE_IMG)

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog
catalog: opm ## Generate catalog manifests and then validate generated files.
	$(OPM) render --output=yaml $(BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/odf.yaml
	$(OPM) render --output=yaml $(ODF_DEPS_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/odf-dependencies.yaml
	$(OPM) render --output=yaml $(OCS_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/ocs.yaml
	$(OPM) render --output=yaml $(OCS_CLIENT_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/ocs-client.yaml
	$(OPM) render --output=yaml $(IBM_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/ibm.yaml
	$(OPM) render --output=yaml $(NOOBAA_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/noobaa.yaml
	$(OPM) render --output=yaml $(CSIADDONS_BUNDLE_IMG) $(OPM_RENDER_OPTS) > catalog/csiaddons.yaml
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
