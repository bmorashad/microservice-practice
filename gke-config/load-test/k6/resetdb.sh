#!/bin/bash

# kubectl exec -n no-mesh-helm -it db-0 -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce; 
kubectl exec -n linkerd-benchmark -it db-0 -c db -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce;
kubectl exec -n istio-benchmark -it db-0 -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce;

http "http://35.240.180.78:8010/reset"
http "http://34.126.89.43:8010/reset"
