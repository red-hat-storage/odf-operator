FROM quay.io/operator-framework/upstream-opm-builder:v1.20.0 AS builder

ARG BUNDLE_IMGS=quay.io/ocs-dev/odf-operator-bundle:latest

RUN opm index add \
--bundles "${BUNDLE_IMGS}" \
--out-dockerfile index.Dockerfile \
--permissive \
--generate

FROM quay.io/operator-framework/opm:latest
LABEL operators.operatorframework.io.index.database.v1=/database/index.db
COPY --from=builder /database/index.db /database/index.db
EXPOSE 50051
ENTRYPOINT ["/bin/opm"]
CMD ["registry", "serve", "--database", "/database/index.db"]
