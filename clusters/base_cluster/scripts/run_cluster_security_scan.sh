#!/bin/bash

if command -v kubescape version; then
    echo "Scanning cluster. This assumes kubescape will find cluster via local kubeconfig."
    echo "NOTE: kubescape is not offline and relies on remote server scan definitions. See kubescape docs."
    # TODO: this is a contrived, partially offline method for downloading a framework, then using it to execute kubescape.
    # However kubescape still interacts with the remote server, and the --keep-local flag appears not to work (likely I'm using it incorrectly).
    # Ideally kubescape would be fully offline and self-contained except for framework definitions, but appears to require a bunch of management.
    kubescape download framework nsa --output .kubescape/nsa.json
    kubescape scan framework nsa --use-from .kubescape/nsa.json
else
    echo "Kubescape is not installed, security scanning not available."
    echo "To install, see: https://github.com/armosec/kubescape"
    exit 1
fi
