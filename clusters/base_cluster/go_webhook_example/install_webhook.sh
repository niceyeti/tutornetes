#!/bin/sh

# Await the mutating-webhook pod before committing the webhook-configuration to activate it.
while $(! kubectl wait -n webhook-example pod -l app=simple-webhook --for=condition=available --timeout=1s); do
    echo "Awaiting mutating-webhook pod..."
done;

kubectl apply -f ./dev/manifests/mutating_webhook.yaml