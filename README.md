# DinD on Argo Workflows

Provides infrastructure definitions to explore and analyze resource utilization when running Docker on Argo Workflow's pods.

## Instructions

1. Provision a k8s cluster that can support `Service` resources with `type: LoadBalancer`. See [Cluster Configurations](#cluster-configurations) for options.

2. Create a namespace and apply the [quick-start-minimal-v3.6.5.yaml](./manifests/argo-workflows/quick-start-minimal-v3.6.5.yaml) manifest:

    ```sh
    kubectl create namespace argo

    kubectl apply \
        --namespace argo \
        --filename ./manifests/argo-workflows/quick-start-minimal-v3.6.5.yaml
    ```

3. Build the [`build-agent`](./build-agent/README.md) Docker image and make it accessible to your k8s cluster:

    ```sh
    docker build \
        --tag build-agent:v0.1.0 \
        --file ./build-agent/Dockerfile \
        .
    ```

4. Start the _success_ workflow:

    ```sh
    argo submit \
        --namespace argo \
        --watch ./manifests/memhog/dind-build-and-run-workflow-success/workflow.yaml
    ```

    This should run for a several minutes.

5. Run the _fail_ workflow to see memory limits in action:

    ```sh
    argo submit \
        --namespace argo \
        --watch ./manifests/memhog/dind-build-and-run-workflow-fail/workflow.yaml
    ```

    This will fail after a few minutes when the docker host sidecar is terminated for consuming too much memory (exit code 137). The exit status is visible in the output of `kubectl` after the workflow fails. Use the command below but replace `xxxxx` with the suffix from the workflow.

    ```sh
    kubectl --namespace argo \
        get pod dind-build-and-run-memhog-workflow-xxxxx -ojson \
        | jq '.status.containerStatuses | map({"image": .image, "name": .name, "state": .state})'
    ```

6. Wait for the _success_ workflow to finish.

    Even though the _fail_ workflow ran and failed due to memory exhastion, the _success_ workflow should still be running and will eventually finish successfully because of the resource limits on each workflow instance.

## Cluster Configurations

### Kind On WSL And Docker Desktop

```text
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
```

**Create a cluster** using `kind create cluster`.

**Support LoadBalancer Services** by running `cloud-provider-kind -enable-lb-port-mapping` in a separate terminal. This application watches the cluster for `Service` resources with `type: LoadBalancer` and handles the networking for them. The `-enable-lb-port-mapping` option creates a container with a port-mapping on the host. This is required on Windows because Docker runs in a VM so the actual load balancer port will be inaccessible.

**Make images available** to the cluster by running `kind load docker-image <image:tag>`. This will push images from the host into the cluster so they can be used by `Pod` resources.
