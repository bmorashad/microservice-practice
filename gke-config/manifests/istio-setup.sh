#!/bin/bash

while getopts n: flag
do
    case "${flag}" in
        n) namespace=${OPTARG};;
    esac
done
if [ -z "$namespace" ]
then
  echo "namespace required: provide it via -n (i.e -n <namespace>)"
  exit
fi

istioctl install --set profile=demo -y
kubectl get crd gateways.gateway.networking.k8s.io &> /dev/null || \
  { kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.6.2" | kubectl apply -f -; }
# kubectl label namespace $namespace istio-injection=enabled
