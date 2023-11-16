#!/bin/bash

kubectl exec -n no-mesh-helm -it db-0 -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce && kubectl exec -n linkerd-mesh-helm -it db-0 -c db -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce && kubectl exec -n istio-mesh-helm -it db-0 -- mysql -uroot -proot -e "truncate table products; truncate table merchantproducts;" ecommerce
