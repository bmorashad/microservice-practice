istioctl install --set profile=demo -y
kubectl label namespace istio-mesh istio-injection=enabled
