FROM quay.io/operator-framework/upstream-opm-builder:v1.20.0 AS builder

ARG BUNDLE_IMGS=quay.io/ocs-dev/odf-operator-bundle:latest

# Copy declarative config root into image at /configs
ADD catalog /configs

RUN opm render --output=yaml ${BUNDLE_IMGS} > /configs/bundle.yaml
RUN opm validate /configs

# The base image is expected to contain
# /bin/opm (with a serve subcommand) and /bin/grpc_health_probe
FROM quay.io/operator-framework/opm:v1.26.0

# Configure the entrypoint and command
ENTRYPOINT ["/bin/opm"]
CMD ["serve", "/configs"]

# Copy declarative config root into image at /configs
COPY --from=builder /configs /configs

# Set DC-specific label for the location of the DC root directory in the image
LABEL operators.operatorframework.io.index.configs.v1=/configs
