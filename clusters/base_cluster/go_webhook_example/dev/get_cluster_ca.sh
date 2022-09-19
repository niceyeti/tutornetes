#!/bin/bash
cat <<- EOF
Note: this script is for obtaining the cluster ca bundle, which is used
as the parameter for the mutating-webhook configuration object's 'caBundle'
parameter. See k8s docs. This could be refactored to live elsewhere.
EOF
bundle=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}')
echo $bundle | fold
echo