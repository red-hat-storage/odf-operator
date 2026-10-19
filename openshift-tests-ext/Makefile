BINARY := odf-tests-ext
GO_PKG_NAME := github.com/red-hat-storage/odf-operator/openshift-tests-ext

.PHONY: build test validate clean

build:
	go build -o $(BINARY) ./cmd/odf-tests-ext/...

test: build
	./$(BINARY) run-suite odf/all

validate: build
	@echo "==> Validating extension info..."
	./$(BINARY) info
	@echo "==> Listing tests..."
	./$(BINARY) list tests --output names
	@echo "==> Listing suites..."
	./$(BINARY) list suites
	@echo "==> Validation passed."

clean:
	rm -f $(BINARY)
