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
#
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -
linkerd viz install | kubectl apply -f -
# kubectl get -n $namespace deploy -o yaml | linkerd inject - | kubectl apply -f -
# kubectl get -n $namespace statefulsets -o yaml | linkerd inject - | kubectl apply -f -
