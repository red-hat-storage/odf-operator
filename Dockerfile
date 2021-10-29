# Build the manager binary
# use our own copy of the official golang image to boost the openshift-ci
# and stop exhausting the docker pull limits.
FROM quay.io/ocs-dev/golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
# cache deps before building and copying source so that we don't need to re-build as much
# and so that source changes don't invalidate our built layer
RUN go install ./vendor/...

# Copy the project source
COPY main.go main.go
COPY Makefile Makefile
COPY hack/ hack/
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/
COPY config/ config/
COPY metrics/ metrics/
COPY console/ console/

# Run tests and linting
RUN make test

# Build
RUN make go-build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
