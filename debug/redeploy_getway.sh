unset http_proxy
unset https_proxy
kubectl delete -f kube_objects/getway-deployment.yml
kubectl apply -f kube_objects/getway-deployment.yml