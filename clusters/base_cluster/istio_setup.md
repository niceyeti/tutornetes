# About

Istio is part of the extended cluster definition and needs to be managed at that level.
Here this just means helm.

## Composing a New Istio-based App

Under istio, in addition to their usual kubes definitions apps consist of gateways, virtual services, and destination rules.

```
$ kubectl api-resources -o wide | grep -i istio
$ kubectl get virtualservices
$ kubectl get destinationrules
$ kubectl get gateway  
```

## Installation and Upgrading

Istio is installed and managed locally (could this be containerized in the future?) by 
downloading it locally to misc/ and adding `istioctl` to PATH.

## Installation

Istio is installed using Helm. See the up.sh script: add the istio helm repo, update it,
install the control plane helm chart, install the istio ingress gateway.

## Upgrading

Download and install the newest Istio version. Add it to path, and update .profile to point
to the new location. Optionally, just remove the version from the path.

## TODO

See these examples:
https://github.com/salesforce/helm-starter-istio/blob/master/ingress-service/values.yaml
https://codeburst.io/istio-by-example-5189edd043da

## Defs


### VirtualService

Recall that K8s Service objects provide a stable network identity for a service via label selector: `app: my_app` and they map ports (a process-level abstraction). A VirtualService is a layer above a Service, and provides additional functionality such as routing, retries, or weighting (e.g. 90/10). VirtualService objects map to Service objects or subsets of Service objects via their `host` definition. Logically these operate as userspace proxies, though I don't know this is the case.

### DestinationRule

DestinationRules are applied after VirtualServices and determine how requests are handled on the receiving end; they implement policies. Direct from the docs:

```
DestinationRule defines policies that apply to traffic intended for a service after routing has occurred. These rules specify configuration for load balancing, connection pool size from the sidecar, and outlier detection settings to detect and evict unhealthy hosts (k8s services) from the load balancing pool.
```

### Debugging Notes

Ingress access: 
* https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/#determining-the-ingress-ip-and-ports
* `kubectl get svc istio-ingress -n istio-ingress`
```
export INGRESS_HOST=$(kubectl -n istio-ingress get service istio-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export INGRESS_PORT=$(kubectl -n istio-ingress get service istio-ingress -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
export SECURE_INGRESS_PORT=$(kubectl -n istio-ingress get service istio-ingress -o jsonpath='{.spec.ports[?(@.name=="https")].port}')
export TCP_INGRESS_PORT=$(kubectl -n istio-ingress get service istio-ingress -o jsonpath='{.spec.ports[?(@.name=="tcp")].port}')
```

Debugging:
1. Examine the Gateway using the steps above.
2. Verify the deployment internally to cluster:
    * kubectl get svc -n dev
    * kubectl exec [dns tools pod name] -n dev -- curl [go app svc name endpoint]


Current sequence for verifying the entire network stack deployment are up:
1. tilt up
2. curl -H "Host: goapp.dev" 172.18.0.4:80/fortune
3. curl -X POST -d "Hello" -v -H "Host: goapp.dev" 172.18.0.4:80/




