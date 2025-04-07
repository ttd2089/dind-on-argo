# DinD on Argo Workflows

Provides infrastructure definitions to explore and analyze resource utilization when running Docker on Argo Workflow's pods.

## Instructions

Provision a k8s cluster that can support `Service` resources with `type: LoadBalancer`.

Create a namespace and apply the [quick-start-minimal-v3.6.5.yaml](./manifests/argo-workflows/quick-start-minimal-v3.6.5.yaml) manifest.

```sh
kubectl create namespace argo
kubectl apply --namespace argo --filename ./manifests/argo-workflows/quick-start-minimal-v3.6.5.yaml
```

Build the [`memhog`](./memhog/README.md) Docker image and make it accessible to your k8s cluster.

```sh
docker build --tag memhog:v0.1.0 --file ./memhog/Dockerfile ./memhog
```

## Cluster Configurations

### Kind On WSL And Docker Desktop

Windows:
    Windows 11 Pro 10.0.26100.3624 Build 26100
WSL:
    Version 2.0.14.0
    Kernel Version 5.15.133.1-1
Docker Desktop:
    Version 4.39.0 (184744)
    Docker Engine v28.0.1
Kubectl:
    Client Version: v1.32.2
Kind:
    kind version 0.27.0
Cloud Provider Kind:
    cloud-provider-kind@v0.6.0

**Provisioning Steps:**

- Create a cluster using `kind create cluster`.
- Ensure `kubectl` context is configured for the kind cluster (`kubectl config set-context kind-kind`).
- Run `cloud-provider-kind -enable-lb-port-mapping` in a separate terminal.
  - While this command is running it will handle the networking for `LoadBalancer` services. The `-enable-lb-port-mapping` option creates a container with a port-mapping on the host. This is required on Windows because Docker runs in a VM so the actual load balancer port will be inaccessible.
