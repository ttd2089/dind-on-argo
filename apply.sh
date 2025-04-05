#!/usr/bin/env sh

kubectl create namespace argo
kubectl apply -n argo -f ./manifests/argo-workflows/quick-start-minimal-v3.6.5.yaml

