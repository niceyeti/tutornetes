![schematic](./schematic.png)

# Golang Kubernetes Webhook Minimal Working Example

This folder contains a minimal working example of a mutating webhook

Webhooks are powerful and simple constructs Kubernetes for modifying k8s objects
on specific events (get, create, patch, etc). See "Kubernetes Best Practices" for an
example, within one of the final chapters on extending kubernetes. Webhooks come in two
flavors, mutating and validating: mutators modify k8s objects, validators accept/reject
k8s objects downstream of mutators. As an side, as expected there are timing issues related
to the use of webhooks, for instance sidecar injection requires that the sidecar image is available,
and multiple webhooks may not obey an order/consistency property. Hence its presumably best
that webhooks implement simple and idempotent behavior.

Webhooks allow some more advanced kubernetes patterns and customizations. For instance, by watching for pod or other objects' creation, one can:
* inject sidecar definitions (currently how Istio injects Envoy proxies)
* enforce security or development rules
* other dynamical behavior, devops flows, and linting

As native constructs, webhooks are much simpler than full CRD/Operator patterns and their code-generation frameworks. Webhooks are easy to comprehend, as a webhook sits in a well-understood location in the k8s object pipeline, and likewise they are implemented using simple http interfaces. This means little to no tooling is required to build up complicated cluster logic, service meshes, and so forth.

# Code

The golang code is in . and in src/. See the mutatePod function; this is where a Pod definition could be modified and returned; currently it just logs a message to indicate the code path has been hit when a pod is deployed. Note that the entire webhook is merely a simple http endpoint, surrounded in kubernetes boilerplate.

# Manual Deployment (vanilla build steps)

1) Build (ensure the image name matches in manifests):
    * docker buildx build -t 127.0.0.1:5000/simple-webhook -f DockerfileDebug .
    * docker push 127.0.0.1:5000/simple-webhook

2) Build the trust info (certs, secret, etc). Cd into dev/, then run 
./cert_gen.sh. There are a few prompts for some manual steps:
    * cd dev
    * ./cert_gen.sh
    * copy the CA bundle in the script output to the caBundle of the mutating-webhook yaml

3) Deploy to cluster, by executing these in order:
    * kubectl apply -f ./dev/manifests/namespace.yaml
    * kubectl apply -f ./dev/manifests/tls_secret.yaml
    * kubectl apply -f ./dev/manifests/deployment.yaml
    * kubectl apply -f ./dev/manifests/service.yaml

4) Await the simple-webhook pod. It may even be advisable to deploy a temporary tools pod to curl the app's /health endpoint, or use the kube api to do so, to confirm the service is running.
    * kubectl get po -l app=simple-webhook

5) Create the mutating webhook configuration.
    * kubectl create -f dev/manifests/mutating_webhook.yaml

6) Verify that the webhook is hit by creating any pod in its namespace:
    * kubectl run busybee -n webhook-example --image=busybox --command -- /bin/sh -c "sleep infinity"
    * kubectl logs [the webhook pod]
    * The logs should show 'hit webhook!' along with the time.

7) Cleanup:
    * kubectl delete ns webhook-example
    * kubectl delete mutatingwebhookconfiguration simple-webhook.acme.com

# Tilt Deployment
First make sure the cluster is free of any previous artifacts using the cleanup steps above.
Then merely run `tilt up` or `up.sh`. To tear down, run `down.sh`. Review the Tiltfile in case of any issues; it was very fragile when written, as enforcing sequential build steps was more difficult than it should have been (IOW, things could be working now by mere chance!).

# Lessons Learned
Developing a webhook is potentially extremely disruptive. Deployed to the wrong namespace, yaml
mistakes, etc., can leave the namespace in a state in which objects like pods cannot
be deployed, because they cannot pass the admission webhook. When developing a webhook,
evaluate the full impact and develop in an environment with minimal collateral damage.
Namespaces, labels, service names, state... they all tend to clash.

# References

* Kubernetes Best Practices, 2020, Chapter 15
* [Slack webhook example code](https://github.com/slackhq/simple-kubernetes-webhook/blob/main/pkg/mutation/inject_env.go)
