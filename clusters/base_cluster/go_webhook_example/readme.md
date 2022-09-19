# Go Kubernetes Webhook Example


This folder contains a minimal working example of a mutating webhook

Webhooks are powerful and simple constructs Kubernetes for modifying k8s objects
on specific events (get, create, patch, etc). See "Kubernetes Best Practices" for an
example, within one of the final chapters on extending kubernetes. Webhooks come in two
flavors, mutating and validating: mutators modify k8s objects, validators accept/reject
k8s objects downstream of mutators. As an side, as expected there are timing issues related
to the use of webhooks, for instance sidecar injection requires that the sidecar image is available,
and multiple webhooks may not obey an order/consistency property. Hence its presumably best
that webhooks implement idempotent behavior.

The power of webhooks is that they allow implementing some more advanced patterns and
customizations. For instance, by watching for pod or other objects' creation, one can:
* inject sidecar definitions (this is how Istio injects Envoy proxies)
* enforce security or development rules
* add other behavior as needed

Webhooks are nice because, as native constructs, they are much simpler than full CRD/Operator
patterns and the code-generation frameworks therein. They are also easy to understand, as a webhook
sits in a well-understood location in the k8s object pipeline, and likewise they are implemented 
using simple http interfaces. This means little if any tooling is required to build up complicated
cluster logic, service meshes, and so forth.

TODO:
* document each programmer use-case and task for mutation:
    * mutating an object
    * rejecting an object (in the mutator, not validation hook)

# Deployment

1) Build (ensure the image name matches in manifests):
* docker build . -t 127.0.0.1:5000/simple-webhook
* docker push 127.0.0.1:5000/simple-webhook

2) Deploy to cluster:
* kubectl apply -f dev/manifests/stack.yaml


Build and push the app:
* docker buildx build -t 127.0.0.1:5000/simple-webhook -f DockerfileDebug .
* docker push 127.0.0.1:5000/simple-webhook
Build the trust info:
* generate certs: ./cert_gen.sh
* copy the CA bundle in the script output to the caBundle of the mutating-webhook yaml
* copy the generated webhook_tls_secret.yaml to stack.yaml
Deploy the app:
1) kubectl create -f dev/manifests/stack.yaml
2) kubectl create -f dev/manifests/mutating_webhook.yaml
3) Verify the webhook is hit by creating any pod:
  * kubectl run busybee -n webhook-example --image=busybox --command -- /bin/sh -c "sleep infinity"
  * kubectl logs [the webhook pod]

Diagnostics:
* Confirm the webhook service is up:
    * Update the dns-tools pod to run in the webhook's namespace
    * kubectl exec dns-tools-78c965764d-84v8t -n webhook-example -- curl -v -k https://simple-webhook.webhook-example.svc.cluster.local/health
* Delete everything, just delete the namespace:
    * kubectl delete ns webhook-example
    * kubectl delete mutatingwebhookconfiguration simple-webhook.acme.com

State:
- After bringing up a new k3d cluster, the caBundle must be in the mutating webhook's yaml.
  See the get_cluster_ca.sh script.



# Lessons Learned
Applying a webhook can be extremely disruptive. Deployed to the wrong namespace, yaml
mistakes, etc., can leave the namespace in a state in which objects like pods cannot
be deployed, because they cannot pass the admission webhook. When developing a webhook,
evaluate the full impace and develop in an environment with minimal collateral damage.
Namespaces, labels, service names, state... they all tend to clash.






# References

* Kubernetes Best Practices, 2020, Chapter 15
* [Slack webhook example code](https://github.com/slackhq/simple-kubernetes-webhook/blob/main/pkg/mutation/inject_env.go)

