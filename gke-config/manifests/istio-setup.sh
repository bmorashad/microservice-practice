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
kubectl label namespace $namespace istio-injection=enabled
