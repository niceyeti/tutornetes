# About

Istio is part of the extended cluster definition and needs to be managed at that level.
Here this just means helm.

Logically Istio is just a layer below microservices (transparent to them) for simplifying their implementation
and separating infrastructure and routing concerns from microservices; in terms of implementation, this is done
by injecting Envoy sidecars beside microservices (or as standalone ingress/egress), which consume configuration
to control request routing.
For the most part, Istio is for defining the graphical definition of a cluster: who talks to whom and how, via
routing, mTLS, request header info, port numbers, and so forth.

## Composing a New Istio-based App

NOTE: Istio is mostly transparent by design, such that apps need not change their code.
The one exception is that for distributed tracing to work, apps must forward headers for Jaeger to successfully capture and identify
specific traces. This may not be the case in the future. See the docs for the requirements.

Under istio, in addition to their usual kubes definitions, apps consist of gateways, virtual services, and destination rules.
Istio is already installed in the cluster at creation (though this claim may be out of date in the future--see up.sh),
and Istio simply extends an app. So the development workflow can be to write the basic kubernetes yaml of any deployment
stack, then to add the following definitions: a 1) Gateway 2) VirtualService 3) DestinationRules (optional).
All the rest are just extensions and advanced features: retries, egress, timeouts, routing, etc., and interfacing to other services.

```
$ kubectl api-resources -o wide | grep -i istio
$ kubectl get virtualservices
$ kubectl get destinationrules
$ kubectl get gateways
```

## Installation and Upgrading

Istio is installed and managed locally (could this be containerized in the future?) by 
downloading it locally to misc/ and adding `istioctl` to PATH.

## Addons

Kiali, Prometheus, et al can be installed via yamls defined in the samples/ folder, using `kubectl apply -f []`.

## Installation

Istio is installed using Helm. See the up.sh script: add the istio helm repo, update it,
install the control plane helm chart, install the istio ingress gateway.

## Upgrading

Download and install the newest Istio version. Add it to path, and update .profile to point
to the new location. Optionally, just remove the version from the path.

## Observability

Adding and running kiali and prometheus:
* WARNING: this is for dev/demo only; these steps are not secure.
* See:
    * https://istio.io/latest/docs/ops/integrations/kiali/#installation
    * https://istio.io/latest/docs/ops/integrations/prometheus/
* kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.13/samples/addons/kiali.yaml
* kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.13/samples/addons/prometheus.yaml
* Find the kiali service port: kubectl describe svc kiali -n istio-system
* Find the kiali pod:
  * `kubectl get pods -n istio-system`
* Port forward to kiali:
  * `kubectl port-forward kiali-699f98c497-wsxv8 8085:20001 -n istio-system`
* Open browser to localhost:8085 then navigate to the graph, set it up to display all namespaces, etc.
* With the go app running, generate some metrics by hitting the app: `curl -H "Host: goapp.dev" 172.18.0.4:80/fortune`
The requests should appear in the kiali graphical display.

## TODO

See these examples:
https://github.com/salesforce/helm-starter-istio/blob/master/ingress-service/values.yaml
https://codeburst.io/istio-by-example-5189edd043da

1. Create cluster with istio and addons
2. Dump and review config
3. Write best-practices doc:
    * define service mesh
    * pros/cons
    * upgrading/updating responsibilities
    * gotchas: header forwarding requires code changes

## Defs

### Architecture
Istio was previously multiple services, since consolidated into a monolith as `istiod`:
* Envoy proxy: sidecar implements secure mTLS between services, retries, failover, health checks, etc., all of the data plane.
* Istiod: the control plane.
    * Pilot
    * Citadel (certs and other secure data)
    * Galley

### VirtualService

Recall that K8s Service objects provide a stable network identity for a service via label selector: `app: my_app` and they map ports (a process-level abstraction). A VirtualService is a layer above a Service, providing additional functionality such as routing, retries, or weighting (e.g. 90/10). VirtualService objects map to Service objects or subsets of Service objects via their `host` definition. Logically these operate as userspace proxies, though I don't know this is the case.

The following rules apply:
* VirtualServices (vs) may mention multiple hosts, but hosts (names) may exist in only one VirtualService.
* The most specific host applies: if vs1 lists host "*.com" and vs2 lists "foo.com" then v2 will match first.
* VirtualServices provide generalized responsibility decomposition: one vs can define the entry to other VirtualServices,
and so on, as needed to divide team responsibilities for individual services.


### DestinationRule

DestinationRules are applied after VirtualServices and determine how requests are handled on the receiving end; they implement policies. Direct from the docs:

```
DestinationRule defines policies that apply to traffic intended for a service after routing has occurred. These rules specify configuration for load balancing, connection pool size from the sidecar, and outlier detection settings to detect and evict unhealthy hosts (k8s services) from the load balancing pool.
```

### Debugging Notes

### Looking at raw config
The Envoy proxies underlying Istio receive configs which can be viewed directly with istioctl:
1. Get the ingress pod name
2. istioctl proxy-config route istio-ingress-69495c6667-jnmnq -o json -n istio-ingress 
3. Or to dump everything: istioctl proxy-config route istio-ingress-69495c6667-jnmnq -o json -n istio-ingress 


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
        * `kubectl exec dns-tools-76d7b69d7-rspbn -n dev -it -c dns-tools -- curl go-app-svc.dev.svc.cluster.local/fortune`
        * `kubectl exec dns-tools-76d7b69d7-rspbn -n dev -it -c dns-tools -- curl go-app-svc/fortune`
Note that these effers are configured with mtls by default, which can be viewed in kiali (look for the lock symbol on edges). So even though curl addresses the k8s service, the request passes through Envoy, thus
tracing the request and implementing mtls transparently.

Current sequence for verifying the entire network stack deployment are up:
0. kubectl get svc istio-ingress -n istio-ingress
1. tilt up
2. Using the ip from (0): curl -H "Host: goapp.dev" 172.18.0.4:80/fortune
3. curl -X POST -d "Hello" -v -H "Host: goapp.dev" 172.18.0.4:80/
4. kubectl exec [dns tools pod] -n dev -- curl -v go-app-svc/fortune

## Istio Course Notes

Sidecar injection: `kubectl label namespace default istio-injection=enabled`.
Best practices for microservices, per istio: architect such that any service can be removed/fail without causing
overall system failure, but rather allow things to continue to function with loss of features.
* Design for failure: define retries and timeouts in all services, in order to preserve overall system determinism.
* Example: test a dense system with a deliberately failing service that times out. Set up the service such that
requests to it fail or timeout, causing overall system degradation. Observe this in kiali on the edge to that service' node.
Then reconfigure the service' istio definition to timeout for requests to that service:
  1. Set the `timeout` field of the VirtualService for the service
  2. Update the service: `kubectl apply -f full-stack-app.yaml`

Envoy: Envoy implements the data plane of Istio, and Istio essentially hides all of Envoy's low-level complexity.
As a demo, Envoy can be run in isolation:
1. See Envoy config here: https://github.com/DickChesterwood/istio-fleetman/blob/master/_course_files/envoy_demo/config.yaml
* The config just defines the indirection of the proxy.
2. Define a Dockerfile to run it:
```
FROM envoyproxy/envoy:latest
RUN apt-get update && apt-get upgrade
COPY config.yaml /etc/envoy.yaml
CMD /usr/local/bin/envoy -c /etc/envoy.yaml
```
SUPER USEFUL: Envoy hosts an admin interface at port 9901 (likely turned off by default in Istio). Using this
you can directly observe request info, stats, and other telemetry; under k8s/Istio this is all exposed via other
application semantics, so its the same info just accessed differently. Still useful to know.

### Kiali, Jaeger, and Grafana

Kiali can be accessed by port forwarding and then interacting with the ui to view what is desired.
* Kiali can generate yaml to modify services and request paths to drop traffic for testing, etc, via the ui.
* Kiali also provides basic VirtualService/DestinationRule linting, albeit via the ui.
* Kiali can write yaml and delete k8s objects, so it should be secured.

NOTE: tracing and other add-on features pose performance risks; for example, Pilot comes with a configurable
parameter for the number of traces to capture. Refer to documentation for examples, such as trace sampling.

Although service meshes operate transparently to app code, the one exception is that Jaeger requires a few headers
(x-request-id, b3, etc) and header forwarding between services for it to identify requests as part of a single trace. See the docs.
The only thing to note is that this does in fact require code changes.
* Envoy will auto-add `x-request-id` to any request lacking this header.
* However, repeated requests inside must also preserve this header, hence the app must do so. Envoy has no way to infer that an incoming and outgoing request are part of the same trace.
* In short: this requires library code to define and forward the headers.

Grafana provides more statistical views of the data, time series.
Grafana provides good monitoring info (cpu, memory, etc).

Use-cases:
* Use these tools to estimate resource usage in development, then use these estimates to define resource limits.
This can be important for production/cloud hosting environments, to avoid running up charges.


### Traffic Management

Canary releases: in standard kubernetes, just use the built in Deployment features for canaries and rollbacks.
* add `version: X` to deployment definitions; kiali will display these versioned service paths on the graph.
* in kiali, you can manually configure request path splitting to send different volumes to different backends
* in istio definitions, weighting is done with VirtualServices
* testing: `while true; curl http://someendpoint.com; sleep 0.5; echo; done;`
* IMPORTANT ARCHITECTURE NOTE: weightings may or may not permit 'sticky' user connections. This depends on whether or not
Istio supports it, but it did not as of 2019.

Example VirtualService with weighted routing:
```
kind: VirtualService
apiVersion: networking.istio.io/v1alpha3
metadata:
  name: some-virtual-service
  namespace: default
spec:
  hosts:
    - my-fancy-service.com  # NOTE: each host def should be an FQDN for a k8s Service
  # note: omitting gateway selector, for clarity
  http:
    - route:
      - destination:  # See DestinationRule for the definition of each of these
        weighting: 10
          host: some-backend-service
          subset: safe  # the name of a DestinationRule subset             
      - destination:
        weighting: 90
          host: some-backend-service
          subset: risky  # the name of a DestinationRule subset
```
Per this example, proxy clients (app code) could reference "my-fancy-service.com", which in turn is 10/90 split across
multiple instances of 'some-backend-service'. This example's 'subset' definitions required that there are corresponding
DestinationRule subset definitions named 'risky' and 'safe' that resolve to the appropriate set of k8s labels
for those pod groupings.

VirtualServices:
* The `name` field is nothing more than k8s info--it is not functional or routable! For example, 
you do not curl the name. By contrast, the `hosts` field is functional.
* A mnemonic for remembering these definitions is virtual-service -> hosts -> route -> destination -> weighting.
* A VirtualService is really just a routing object.
* It is NOT a replacement for a k8s Service; it is an extension to k8s Services to provide client-side
indirection / routing via the envoy proxies and ingress gateways.
* It is important to visualize the request path for these properties: vanilla Services provide basic stable identities for
sets of underlying Pods; VirtualServices sit in front of one or more Services or even external services.
Services are implemented in the kernel by iptables; VirtualServices are userspaces proxies implemented by Envoy.
* Pilot <----> VirtualServices
* Kube-Proxy <----> Services
* VirtualServices are defined as 'routes' in envoy, viewable using: `istioctl proxy-config route istio-ingressgateway_PODNAME -o json`
 
DestinationRules: these provide a layer of indirection for the `destination` fields of VirtualServices,
as defined in the yaml above. They provide load-balancing rules: given this virtual service or request,
to which pods it should be routed.
* load balancing
* pod grouping
```
kind: DestinationRule
apiVersion: networking.istio.io/v1alpha3
metadata:
  name: some-service   # This name can be anything; ie, is not functional / routable.
  namespace: default
spec:
  host: some-backend-service
  trafficPolicy: ~
  subsets:   # NOTE: these are SELECTORS
    - labels:   # selector
        version: safe   # find all pods labeled with 'version=safe'
      name: safe
    - labels:
        version: risky  # find all pods labeled with 'version=risky'
      name: risky
```
ServiceEntries: understand the relationship with ServiceEntries vs native K8s services and endpoints. ServiceEntries are useful for things like external services; they provide external services a definition to which things like DestinationRules can be applied.

## Load Balancing
Per the above, VirtualServices and DestinationRules provide indirection to subsets of a service.

Sticky sessions: traditionally, sticky sessions could not be implemented using weightings (though this may have changed
since 2019). They can be configured using 'consistent hashing'.
Consistent hashing: 'if hash is even, do to pod 1; if hash is odd, go to pod 2' etc.
Configured in DestinationRules:
```
kind: DestinationRule
apiVersion: networking.istio.io/v1alpha3
metadata:
  name: sticky-sessions
spec:
  host: my-service.svc.cluster.local
  trafficPolicy:
    loadBalancer:
      consistentHash:
        httpCookie:
          name: user
          ttl: 0s
```
Load balancing can be implemented by defining the parameters of a hashing algorithm based on 
request info (source ip, http headers, cookie, etc).
NOTE: hashing still does not provide sticky sessions, since weighing rules are applied upstream of hashing / lb rules!
* consider weighting and sticky sessions to be incompatible patterns

One hazard of consistent-hash is that some hashing parameters must be forwarded across whatever chain of microservices
they traverse; a common error seems like it would be to pass a header with curl, hash it, but get unexpected 
results because the header is not forwarded to the target service.
* One novel solution is for your client web app to implement a special header prefix, and to forward any header with that
prefix across the system (e.g. in some library code). Seems like a smelly pattern... just spitballing.

Patterns that consistent hashing supports: forward specific requests to the same backing deployments.
This should ONLY be used for things like canaries, and not assumptions within apps themselves, or at least,
only very, very thoughtfully. For example, this pattern supports better caching of user queries/info within pods,
but this assumption breaks the rule of 'cattle not pets' in microservices, and the feature was not designed for it.
On the other hand, the performance enhancement of maintaining 'sessions' using hashing could implement a simple optimization
if done transparently with respect to the app; in the real world, this almost immediately implies abuse by management
and scope bloat to link apps to that dependence somehow, breaking a core responsibility of microservices.

In summary: consistent hashing is useful, but the pattern must be implemented conscientiously and should preserve 12-FA principles.

## Ingress Gateways
One important lesson of this course is that Istio object--Gateways, VirtualServices, DestinationRules, consistent hashing--provide
a hierarchy by which to implement certain requirements. For example, to implement an experimental canary release of software,
one might use a VirtualService with weighting, or load balancing based on Destination rules. However this might be too low-level.
A higher level option is to implement this using a Gateway object.

Distinguish physical from istio Gateway: a Gateway is a definition for rules binding to hosts (virtual services, usually). The physical gateway (e.g. istio-ingress) is separate, and may be selected by one or more Gateway objects.

Using an ingress-gateway and defining a Gateway+VirtualServices ensures that traffic traverses the envoy proxies.
Of course, one could (ie, in dev) implement Services with NodePort, but this common mistake means that your traffic will
not traverse the Envoy proxies, and hence no istio features/rules applied. The point is simply to distinguish how
your requests are traveling, vs vanilla k8s.
* Reconfigure your /etc/hosts file to map to an ingress svc as such:
  * find the svc external port of the istio-ingress service: `kubectl get svc -n istio-system -o wide`
  * Add its ip to /etc/hosts: `172.1.2.3 my-webapp.com`

## Routing
In VirtualServices, define prefix-based routing as:
```
# ... in some VirtualService spec definition
http:
  - match:
    - uri:  # IF
      prefix: "/path-one"
    - uri:  # OR
      prefix: "/path-1"
    route:  # THEN
      destination:
        host: my-webapp
        subset: some-subect # defined in some DestinationRule
```
Of course this also enables rewriting, and of course, this may be extremely problematic and buggy given that the webapp's
paths may be inconsistent.
* infrastructural and webapp internal path logic must be consistent
* ensure this is the case, or requests will crap out

Other routing logic may be implemented based on: prefix, headers, port, query params, etc.

## Full Canary Example
Canaries can be implemented in two separate VirtualServices, each with its own matching requirements or subdomains.
Another virtual service could sit in front of them, and applying routing/weighting and so forth.
Thus, VirtualServices provide a generalization of indirection via recursion: the first layer virtual service applies weighing,
the virtual services to which it forwards implement their own logic (routing, etc), and so on, in a layered manner.

## Fault injection
Again, the workflow for generating yaml can be done using kiali to generate and apply some rules, copy the relevant
fields of the yaml, then selecting 'delete custom rules' on the ui, then applying the manually edited yaml derived
from the kiali-generated yaml.
* failures
* delays
Fault injection allows chaos testing across a system: generate fault yaml for a service, deploy it, send requests, analyze results.

## Circuit Breaking
Circuit breaking is intended to prevent cascading failures, whereby one service's failure cascades to the others,
usually via timeouts or connection exhaustion as we saw at SEL.
Circuit breaking can be embedded into client code libraries as implementations which simply track successful/failed
requests and quit requests based on some configuration.
A "Fail Fast" mechanism to prevent saturation of resources: connections, buffers, pending requests at client/server, etc.

Note: with Istio/Envoy, code-based circuit breakers are a bit of an antipattern, since Istio implements them at a layer
that is easier to test, undo, trigger, or close. This isn't possible with code circuit breakers since code changes would
require the service to restart. On the other hand, there is no reason not to define circuit breaking in software,
especially simply using queues where a service simply stops requesting when its request queue is full. Just be 
mindful of the benefits of defining circuit breaking at the infrastructure level.


## mTLS and Intracluster Security
Mutual encryption is needed for scenarios such as when services run on separate nodes,
and additionally block ALL unsecure traffic.

mTLS must be understood in the context of a container making a request "http://some-service:80",
which then goes to the Envoy proxy, and thus through the service mesh layers. So it is CRITICAL
to understand how requests are transformed, from the code context (the app developer's view) to
the service mesh context.

mTLS is now implemented by default in Istio; however you can configure it as STRICT or PERMISSIVE 
to gradually upgrade a cluster.

## Istioctl

Istioctl must be installed and added to path: `export PATH="/some/dir:$PATH"`
One of the most important tasks is using istioctl to define profiles:
1. * `istioctl profile dump demo > raw_demo_settings.yaml`
  * `istioctl profile dump default > raw_default_settings.yaml`
2. Review and edit the profile
3. Apply the new settings: `istioctl install -f some_settings.yaml`

Commands:
* istioctl 
* `kubectl -n istio-system get IstioOperator installed-state -o yaml > installed-state.yaml`
* Generate istiod k8s yaml: `istioctl manifest generate input.yaml > some_k8s.yaml`
  * This is not suggested; instead use the builtin istioctl native definitions and istioctl api
* istioctl proxy-status: use this to view proxy status and istiod daemon connections

## Upgrade Best Practices

Upgrading is highly complex and potentially error prone, due to the requirements of live systems maintaining continuity.
Like our experience at CNL, I wouldn't even suggest this; upgrades never work, except with direct access to the system
and hours of direct debugging and TONS of (mostly meaningless) experience. Instead, build a workflow for 
rebuilding from scratch.

Istio allows two kinds of upgrading:
1. Canary upgrade of the control plane: least risk, more complex.
2. In-place upgrades: simplest but riskiest method of upgrading.
    * istioctl upgrade

Gotchas:
* The famous sidecar bug: Envoy sidecars are not guaranteed to start before app containers, often causing 
issues when apps expect some resource to be ready.
  * Resolution: set `values.proxy.holdApplicationUntilProxyStarts`
* You cannot curl VirtualServices from other apps, since VirtualServices do not create DNS entries. Think simply,
and instead curl the k8s service directly.
* The istioctl tool is versioned, so this process also involves selecting the istioctl version, e.g. istioctl17
for the 1.7 tool. Just be aware of this as a task.
* Pods, and in turn the Envoy proxy sidecars, are immutable, so upgrading requires restarting pods to get new proxies.
    * Run `kubectl delete pod [pod]` on each old pod
    * Also possible, restart deployments: `kubectl rollout restart [deployment]`
* Upgrading distributed system components nearly always incurs hidden dependencies and issues when transitioning to new state.
  * upgrading is a low visibility operation
  * proxies detect and transition to new istiod daemon, but the assumptions therein not clear or an implementation detail.
* There is a best-practice of labeling the namespace with the revision version, by which proxies will remain connected
to the previous version and then new pods stood up gradually and independently. Just be aware.

Use Grafana to view and navigate/understand istio components and versions: -> istio mesh dashboard.

### Canary upgrades
This is a more complex process but lower risk and with greater visibility. Again, avoid if possible, see docs if needed.
This is just for learning:
1. Label the namespace `istio.io/rev: 1.7`. This will preserve existing proxy connections to the istiod (here, 1.7)
2. Install the new version: `istioctl install --set profile=demo --set revision=1-8`
  * Result: istiod 1.7 and 1.8 are running, but the previous system still only connects to 1.7 daemon.
3. Restart pods or deployments manually; when up, they will see the new revision label, and the proxies will search for this istiod version.
Source: https://istio.io/latest/docs/setup/upgrade/canary/

### Live cluster switchover
The highest-cost, least risk approach: 
1. Create an entirely new cluster, possibly still connecting to the same db backend as the previous istio-versioned cluster.
2. Redirect your load balancer to the new cluster, or do so using dns.
3. Drain the old cluster, for graceful transition to the new system.

### Best practices and interesting bits

Istio exemplifies the use of labels for performing various forms of behavior or system identification: sidecar-injection=enabled, and so on.
In pure k8s, labels drive command line queries and selectors, for example. Yaml definitions can even have more complex 
labeling semantics using matchExpression, but whatevs.
Some k8s commands for refreshment:
* `kubectl get pods --show-labels`
* `kubectl get pods -l environment=production,tier=frontend`
* `kubectl get pods -l 'environment in (production, qa)'`
* `kubectl label pod -n some-namespace mypod some-label=true`
For example, labels could be drive a CI/CD pipeline or other system.
Labels and annotations provide the most straightforward state storage mechanism.

* Consistent hashing and weighting rules are incompatible
* When implementing features like canaries, consider the Gateways, VirtualServices, and DestinationRules as pieces of a hierarchy,
each with their own tradeoffs. Per usual, provide multiple possible implementations, e.g. five solutions, and rank them by the features
they do/dont extend or allow.

* All clusters should be regeneratable using source control.
* Repeating, perhaps, but all state should be committed to source control.
* Pods should be cattle, not pets: this abets many patterns and independent flexibility by which to make changes to a system.


Checkout the Fallacies of Distributed Computing:
* https://en.wikipedia.org/wiki/Fallacies_of_distributed_computing


### Course revisions and feedback

* Use k3d, not minikube. Kind may also work well.
* istio-ingressgateway is not just istio-ingress