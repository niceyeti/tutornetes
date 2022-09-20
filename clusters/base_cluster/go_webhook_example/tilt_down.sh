#/bin/sh
# merely adds '--delete-namespaces' to 'tilt down' because its tedious...
tilt down --delete-namespaces
kubectl delete mutatingwebhookconfiguration simple-webhook.acme.com
kubectl create -f ./dev/manifests/namespace.yaml