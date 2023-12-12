#!/bin/bash

# kubectl exec -n no-mesh-helm -it db-0 -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce; 
kubectl exec -n linkerd-benchmark -it db-0 -c db -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce;
kubectl exec -n istio-benchmark -it db-0 -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce;

linkerd=$(kubectl get ing -A | tail -n +2 | rg "linkerd" | awk '{print $5}')
istio=$(kubectl get ing -A | tail -n +2 | rg "istio" | awk '{print $5}')


http "http://$linkerd/reset"
http "http://$istio/reset"

./selectproducts.sh
