#!/bin/bash

./clean.sh
make docker-build
make docker-push
make install
make deploy
echo "Sleeping a sec for controller to start (takes about 15s)..." && sleep 15
kubectl create -f temp/test_goop.yaml
kubectl logs -n goop-system -l app=goop-controller-manager -f
