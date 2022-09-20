#/bin/sh
# merely adds '--delete-namespaces' to 'tilt down' because its tedious...
tilt down --delete-namespaces

# Delete anything remaining, if it still exists.
# `tilt down` should remove everything cleanly, but in my experience the tool has a
# hard time synchronizing k8s object deconstruction. The specification by which
# it does so (a combination of the tilt file api and starlark language) is not
# even clear as to how the inference of destruction is performed.
kubectl delete mutatingwebhookconfiguration simple-webhook.acme.com
kubectl delete -f ./dev/manifests/namespace.yaml