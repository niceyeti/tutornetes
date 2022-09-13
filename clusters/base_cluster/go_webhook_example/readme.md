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

# References

* Kubernetes Best Practices, 2020, Chapter 15
* [Slack webhook example code](https://github.com/slackhq/simple-kubernetes-webhook/blob/main/pkg/mutation/inject_env.go)

