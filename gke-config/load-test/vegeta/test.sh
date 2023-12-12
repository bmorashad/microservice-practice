#!/bin/bash

if [ "$1" = "linkerd" ]; then
  meshName="linkerd"
else
  meshName="istio"
fi
ing=$(kubectl get ing -A | tail -n +2 | rg "$meshName" | awk '{print $5}')
mesh=$(kubectl get ing -A | tail -n +2 | rg "$meshName" | awk '{print $1}' | rg ".*" -r "$meshName")
resultsDir=$meshName-results

run_test() {
  now=$(date +'%d-%m-%y-%H:%M')
  testMeta=${1}w.${2}mw.${3}rps.${4}d
  mkdir -p $resultsDir/$testMeta
  resultsFileName="$resultsDir/$testMeta/$mesh.results.$testMeta:$now.bin"
  attackName=$mesh.$testMeta
  echo "GET http://$ing/create-products/random
  GET http://$ing/products" | vegeta attack -name=$attackName -workers $1 -max-workers $2 -rate $3/1s -duration=$4 | tee $resultsFileName | vegeta report
}

for ((i = 0; i < 10; i++)); do
  run_test 10 64 100 1m
done

for ((i = 0; i < 10; i++)); do
  run_test 50 64 100 1m
done

for ((i = 0; i < 10; i++)); do
  run_test 50 150 100 1m
done

for ((i = 0; i < 10; i++)); do
  run_test 50 200 150 1m
done

for ((i = 0; i < 10; i++)); do
  run_test 10 0 0 1m
done
