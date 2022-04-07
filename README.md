# OpenShift Data Foundation Operator

This is the primary operator for Red Hat OpenShift Data Foundation (ODF). It
is a "meta" operator, meaning it serves to facilitate the other operators in
ODF by providing dependencies and performing administrative tasks outside
their scope.

## Deploying pre-built images

### Installation
The ODF operator can be installed into an OpenShift cluster using Operator
Lifecycle Manager (OLM).

For quick install using pre-built container images.

```
make deploy-with-olm
```

This creates:

* a custom CatalogSource
* a new openshift-storage Namespace
* an OperatorGroup
* a Subscription to the ODF catalog in the openshift-storage namespace

You can check the status of the CSV using the following command:

```
oc get csv -n openshift-storage
```

This can take a few minutes. Once PHASE says Succeeded you can create a
StorageSystem.

StorageSystem can be created from the console, using the StorageSystem creation
wizard. From the CLI, a StorageSystem resource can be created using the example
CR as follows,

```
oc create -f config/samples/ocs-storagecluster-storagesystem.yaml
```

## Development

### Build

#### ODF Operator

The operator image can be built via

```
make docker-build
```

#### ODF Operator Bundle

To create an operator bundle image with the bundle run

```
make bundle-build
```

#### ODF Operator Catalog

An operator catalog image can then be built using

```
make catalog-build
```

### Deploying development builds

To install own development builds of ODF, first build and push the odf-operator
image to your own image repository.

```
export REGISTRY_NAMESPACE=<quay-username>
export IMAGE_TAG=<some-tag>
make docker-build docker-push
```

Then build and push the operator bundle image.

```
export REGISTRY_NAMESPACE=<quay-username>
export IMAGE_TAG=<some-tag>
make bundle-build bundle-push
```

Next build and push the operator catalog image.

```
export REGISTRY_NAMESPACE=<quay-username>
export IMAGE_TAG=<some-tag>
make catalog-build catalog-push
```

Now create a ODF operator and follow the [Installation](#installation)

```
export REGISTRY_NAMESPACE=<quay-username>
export IMAGE_TAG=<some-tag>
make deploy-with-olm
```

## Running Unit test

Unit tests can be run via

```
make test
```

To run a single test

```
go test -v github.com/red-hat-storage/odf-operator/controllers \
    -run TestIsVendorSystemPresent
```

## Contribution

To contribute to the project follow the [contribution](CONTRIBUTING.md) guide.
