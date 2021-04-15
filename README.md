# Bigtable Autoscaler Operator
This is a Kubernetes operator to autoscale Bigtable

## Development cycle

### Option 1: Run with Tilt

Tilt is tool to automate development cycle and has features like hot reloading.

1. Install [kubebuilder](https://book.kubebuilder.io/quick-start.html#installation).
2. Install tilt (follow the [official instructions](https://docs.tilt.dev/install.html)). Make sure you create a kuberenetes cluster with ctlptl and kind as instructed.
3. Execute `tilt up` to run the operator on your local cluster.

### Option 2: Run manually

To build with docker run:

#### Provide a development cluster

You can use kind to run a sample cluster

```sh
kind create cluster
```

check that your cluster is correctly running

```sh
kubectl cluster-info
```

#### Apply Custom Resource Definition
```sh
make install
```

#### Build docker image with manger binary
``` sh
make docker-build
```

#### Load this image to the cluster
```sh
kind load docker-image controller:latest
```

#### Run the manager on the cluster
```sh
make deploy
```

#### Check pods and logs
```sh
kubectl -n bigtable-autoscaler-system logs $(kubectl -n bigtable-autoscaler-system get pods | tail -n1 | cut -d ' ' -f1) --all-containers
```

## Running tests
```sh
go test ./... -v
```

# Usage

#### Apply some autoscaler
```sh
kubectl apply -f config/samples/bigtable_v1_bigtableautoscaler.yaml
```

#### Create secret
```sh
kubectl create secret generic bigtable-autoscaler-service-account --from-file=service-account=./your_service_account.json
```

#### Create role and rolebinding to read secret
```sh
kubectl apply -f config/rbac/secret-role.yml
```


