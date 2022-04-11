#!/bin/bash

if command -v kubescape version; then
    echo "Scanning app. Note: this assumes kubescape will find cluster via local kubeconfig."
    kubescape download framework nsa --output .kubescape/nsa.json
    helm template go_app/chart --values go_app/chart/values.yaml --dry-run | kubescape scan framework nsa --use-from .kubescape/nsa.json -
else
    echo "Kubescape is not installed, security scanning not available."
    echo "To install, see: https://github.com/armosec/kubescape"
    exit 1
fi
