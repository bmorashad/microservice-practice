#!/bin/bash

echo "[Istio Products Count]"
kubectl exec -n istio-benchmark -it db-0 -c db -- mysql -uroot -proot -e "select count(*) from products" ecommerce
echo "[Linkerd Products Count]"
kubectl exec -n linkerd-benchmark -it db-0 -c db -- mysql -uroot -proot -e "select count(*) from products" ecommerce
