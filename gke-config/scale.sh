#!/bin/bash

if [ "$1" = "down" ]; then
  gcloud container clusters resize cluster-1 --num-nodes=0 --region=asia-southeast1 --node-pool=system --quiet
  gcloud container clusters resize cluster-1 --num-nodes=0 --region=asia-southeast1 --node-pool=istio-meshed-big --quiet
  gcloud container clusters resize cluster-1 --num-nodes=0 --region=asia-southeast1 --node-pool=linkerd-meshed-big --quiet
  gcloud container clusters resize cluster-1 --num-nodes=0 --region=asia-southeast1 --node-pool=no-meshed-big --quiet
else
  gcloud container clusters resize cluster-1 --num-nodes=1 --region=asia-southeast1 --node-pool=system --quiet
  gcloud container clusters resize cluster-1 --num-nodes=1 --region=asia-southeast1 --node-pool=istio-meshed-big --quiet
  gcloud container clusters resize cluster-1 --num-nodes=1 --region=asia-southeast1 --node-pool=linkerd-meshed-big --quiet
  gcloud container clusters resize cluster-1 --num-nodes=1 --region=asia-southeast1 --node-pool=no-meshed-big --quiet
fi
