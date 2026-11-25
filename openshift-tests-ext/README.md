# ODF OpenShift Tests Extension (OTE)

This directory contains the OpenShift Tests Extension (OTE) binary for ODF testing.

## Overview

The `odf-tests-ext` binary implements the OTE v1.1 protocol so ODF checks can run under `openshift-tests` with `--extension-local-binaries`.

## Building

```bash
make build
```

## Running Tests

```bash
./odf-tests-ext run-suite odf/all
./odf-tests-ext run-suite odf/smoke
./odf-tests-ext list tests
./odf-tests-ext info
```

## Test Suites

- `odf/all`
- `odf/smoke`

Scripts are embedded in the binary. The adapter extracts them to a temp directory at runtime and sets `INTEROP_SCRIPTS_DIR`. Shell scripts use exit `2` for skip conditions; the adapter maps that to OTE `skipped`.

## Related

- JIRA: https://redhat.atlassian.net/browse/INTEROP-9590
- OpenShift Tests Extension framework: https://github.com/openshift-eng/openshift-tests-extension
