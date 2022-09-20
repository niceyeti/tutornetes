#!/bin/sh

# Silly, but the sole purpose of this script is to implement something extremely basic,
# but opaque or Rube-Goldbergy in tilt/starlark: ensure the namespace exists before
# everything else is built. Nothing in the docs or in numerous github issue clearly
# spells how to make build-order dependencies explicit for basic (non-resource) k8s objects.
kubectl create -f ./dev/manifests/namespace.yaml
tilt up
