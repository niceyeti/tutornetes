#!/bin/bash

cat <<-EOF
-----------------------------------------------------------------------------------------------
Generating tls cert and prv key; these is purely for development.
The parameters within this script are the desired CN and subjectAltName
in DNS (the k8s service FQDN, including namespace), per this mapping:

    CN: the app name as required per tls
    Namespace of k8s service: this is required for the FQDN
    subjectAltName: DNS FQDN of the k8s service, e.g. 'simple-webhook.default.svc.cluster.local'

The corresponding k8s tls secret is generated and output to /dev/manifests/.

NOTE: this script only needs to be run once and the outputs managed as described
before being committed.
-----------------------------------------------------------------------------------------------
EOF
echo
read -p "Press ENTER to continue"
echo

openssl genrsa -out ca.key 2048

openssl req -new -x509 -days 365 -key ca.key \
  -subj "/C=AU/CN=simple-webhook"\
  -out ca.crt

openssl req -newkey rsa:2048 -nodes -keyout server.key \
  -subj "/C=AU/CN=simple-webhook" \
  -out server.csr

openssl x509 -req \
  -extfile <(printf "subjectAltName=DNS:simple-webhook.webhook-example.svc") \
  -days 365 \
  -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt

echo
echo ">> Generating kube secrets..."
kubectl create secret tls simple-webhook-tls \
  --cert=server.crt \
  --key=server.key \
  -n webhook-example \
  --dry-run=client -o yaml \
  > ./manifests/webhook_tls_secret.yaml

echo
echo ">> MutatingWebhookConfiguration caBundle:"
cat ca.crt | base64 | fold

echo
read -p "Copy the caBundle above into the mutating webhook config yaml, then press ENTER and have a nice day."
echo

rm ca.crt ca.key ca.srl server.crt server.csr server.key