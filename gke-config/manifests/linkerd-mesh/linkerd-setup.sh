linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -
kubectl get -n linkerd-mesh deploy -o yaml | linkerd inject - | kubectl apply -f -
kubectl get -n linkerd-mesh statefulsets -o yaml | linkerd inject - | kubectl apply -f -
