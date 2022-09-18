
# Productivity

## Config and command line conveniences

Config files define clusters, users, and contexts.
* cluster: a kubernetes cluster, i.e. an api server
* user: users can be created and configured per certificate files for authenticating to different clusters
* context: contexts are a user+namespace+cluster
    * `kubectl config --kubeconfig=config-demo set-context dev-frontend --cluster=development --namespace=frontend --user=developer`
    * `kubectl config --kubeconfig=config-demo use-context exp-scratch`

Alias kubectl as:
* alias k=kubectl

-A is shorthand for --all-namespaces:
* kubectl get pods -A

Show all of the api-resources and CRDs:
* `kubectl api-resources`
* `kubectl explain deployment`
* most commands have a '-h' option as well

## Dry-run
Use dry-run to create starter yaml as follows:
* `kubectl create deployment my-dep --replicas=2 --image=busybox --dry-run=server -o yaml`
  * `dry-run=client` directive generates yaml only on the client
  * `dry-run=server` passes the command to the server to generate spec, after applying all webhooks, etc.

## Diff
Use the diff command against live objects to check the differences before applying them:
* `kubectl diff -f some_object.yaml`

## Watch
Use --watch to watch for events per some object.
* `kubectl get pod my-friggin-pod -o wide -n dev --watch`

## Shorty nix commands
Use while loops to demo issues:
* `kubectl run testpod -n dev --image=busybox --restart=Never --command="/bin/sh -c 'while true; do curl -v localhost/some-service; done;'`


# Primary verbs
| Command | Def |
| --- | --- |
| create | create a new object |
| delete | delete an object
| get | get objects |
| diff       |   Diff the live version against a would-be applied version  |
| apply      |   Apply a configuration to a resource by file name or stdin  |
| patch      |   Update fields of a resource  |
| replace    |   Replace a resource by file name or stdin  |


# Pod definitions
k create pod my-pod --image=busybox --restart=Never --command env



# Labels and Annotations
Note: labels and env vars are defined as csv when passed imperatively on the command line.

* kubectl get pods -l env=prod
* k get pod my-pod --show-labels
* k run my-pod --image=busybox --restart=Never --labels="env=dev,tier=db" --env="DB_HOST=$DB_HOST"
* k label pod my-pod env=dev

Now delete the label by key:
* k label pod my-pod env-

Annotation commands are virtually the same as for the label command:
* kubectl annotate pod my-pod some-key=blahblah
* kubectl annotate pod my-pod some-key-

# Imperative networking

* Containers within a pod share the same network namespace, and thus the same localhost.
But this is deeper since an app may bind to port 80 within a container, and yet have non runtime
privileges outside of the container to bind to low numbered ports. Thus a container can bind to 80
inside, and map to port 8080 on the outside; inside the container the process behaves as if it has these privileges, but once outside (8080) it does not. Think of this as a trust boundary.
* Map a host port to a cluster-internal pod on the 10.* network:
  * `kubectl port-forward my-pod -n dev 8080:80`
  * This effectively makes the api server act as a gateway to a pod, deployment, or service.
* Proxy the local system to the api server:
  * `kubectl proxy --port=8080`
  * `curl http://localhost:8080/api/`
  * Exposing the api-server REST API allows for direct and perhaps more efficient testing (see next section)
* Every pod receives the server cert, token, and namespace by which to communicate directly with
the api server. The api-server location params are provided as env vars to every pod.
  * by default, service account secrets are in /var/run/secrets/kubernetes.io/serviceaccount
    * `CA_CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`
    * `TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)`
    * `curl -v --cacert $CA_CERT -H "Authorization: Bearer $TOKEN" https://$KUBERNETES_PORT/`
    * `curl -v --cacert $CA_CERT -H "Authorization: Bearer $TOKEN" https://10.43.0.1:443/api/v1/namespaces/dev/pods`
    * As one-liner:
        * `CA_CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt; TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token); curl -v --cacert $CA_CERT -H "Authorization: Bearer $TOKEN" https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT/api/v1/namespaces/dev/pods`
    * Note: interacting with the api-server requires configuring roles/role-bindings to allow specific users to CRUD object
* Use `expose` to create Services on the fly, again purely for debugging:
  * `kubectl expose pod simple-go-app-64b5c755fc-fkwwx -n dev --port=4444 --target-port=80 --name=my-awesome-svc`
  * This service may then be curl'ed from another container:
    * `kubectl exec -it -n dev dns-tools-1234 -- /bin/sh`
    * `# curl my-awesome-svc.default.cluster.local:4444/fortune`

# Ultimate Kommand Line 

First off, get acquainted by aliasing the command and checking the current context:
* alias k=kubectl
* k config view
* Use `--help` to explain any command or its parameters
* Use `kubectl api-resources` to display and understand k8s objects and CRDs

Tools like k3d will autopopulate the kube config, located at $KUBECONFIG or the default at
`~/.kube/config`. In development, it is imperative to understand how your user is configured,
etc., or it may bite you. The definition is found using `k config`, but the dev framework 
will be the initializer.

## Config 

The kubeconfig contains info on the current configuration.
The context is defined as a cluster and user.

* kubectl config view
* kubectl config 

# Getters

Useful getters:
* kubectl get pods -n dev
* kubectl get pods -A --show-labels
* kubectl get pod mypod -o yaml -n dev

Get pods with a label, or for specific key/val:
* kubectl get pods -l env
* kubectl get pods -l app=frontend

Output labels as columns:
* kubectl get pods -L env

Show extended information and status:
* kubectl describe pod mypod

Check docs for objects:
* kubectl explain [object]
* kubectl explain pod
* kubectl explain netpol

Field selectors can be used to drill into yaml specs using the normal yaml keys:
* kubectl get secret -n default -o yaml --field-selector=metadata.name=mysecret

# Formatting

Formatting is just a part of getters, but deserves its own section since it defines consumption by other users. Output formatting is done using jsonPath.

Append certain columns/labels to columnar output:
* kubectl get pod -A -L env

Output only certain info from specs:
* kubectl get pods -A -o json
* kubectl get pods -A -o=jsonpath='{@}'
* kubectl get pods -A -o=jsonpath='{.items[0]}'
* kubectl get pods -A -o=jsonpath='{.items[0].metadata.name}'
* kubectl get pods -A -o=jsonpath="{.items[*]['metadata.name', 'status.capacity']}"
* kubectl get pods -o=jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.startTime}{"\n"}{end}'

Show all pods' volumes:
* kubectl get pods -A -o=jsonpath='{.items[*].spec.volumes}'

# Logging
Logging should really be performed by additional infrastructure and exported/queried as such. Default k8s logging captures stdout of procs, which is useful for limited prototyping an dev.

Get logs:
* kubectl logs mypod

Get the previous logs of a restarting container/pod:
* kubectl logs mypod --previous

Get logs from a specific container in a pod:
* kubectl logs mypod -c sidecar

# Labels and Annotations

Add or overwrite a label:
* kubectl label pod mypod env=prod --overwrite

Annotations are the same:
* kubectl annotate pod mypod oncall="555-555-5555" --overwrite

# CRUD

Create starter yaml for objects:
* kubectl run somepod --image=busybox --command --dry-run=client -o yaml > test.yaml -- /bin/sh -c "sleep infinity"
* kubectl expose pod somepod -n dev --port 8080 --dry-run=client -o yaml

Delete objects:
* kubectl delete [object type] [object name]
* kubectl delete pod mypod
* kubectl delete pod mypod --grace-period=0

Modifying objects can be done by several workflows, just pick your poison:
* kubectl apply -f spec.yaml
* kubectl edit pod mypod
* kubectl get pod mypod -n dev -o yaml > spec.yaml && gedit spec.yaml
    * kubectl apply -f spec.yaml
The most native workflow is probably `kubectl edit`. `kubectl apply` requires that the object was created using `kubectl apply` or `kubectl create --save-config`, so isn't as smooth as you might think. `--save-config` stores the object's definition in the pod's annotation.

# Validation

Oftentimes you want to validate a yaml definition before actually submitting it, or generate starter yaml and edit it before submitting it.

Use --dry-run=client or --dry-run=server to generate client/server yaml. The difference is that (I'm assuming) the `--dry-run=server` will actually put the object through the entire admission mutation/validation pipeline without applying it, thus the yaml will be as complete as possible. So `--dry-run=client` is good in a pinch, but `--dry-run=server` is preferable since it is strictly more complete.
* kubectl run mypod --image=nginx -n dev --dry-run=server

Validation can be done similarly, using the `--validate=true` or `--validate=server` (these are equivalent). See the full description of both using `--help`.
1) kubectl run mypod --image=busybox --dry-run=server -n dev --port=80 --expose --command -o yaml > test.yaml -- /bin/sh -c "sleep infinity"
2) kubectl create -f test.yaml --validate=true --dry-run=server

Note that both use `--dry-run`. (1) uses it to generate yaml, whereas (2) ensures the object is validated completely, as if it went through all admission controls, without actually being created.

# Running

Creating and running imperatively is a lot of wizardry and beautiful one-liners:
* kubectl run mypod --image=nginx --port 80 --expose
* kubectl run toolz --image=busybox -it --rm --command -- /bin/sh

# Diagnostics

Some handy diagnostics commands:
* kubectl exec toolz-pod -- curl some-service.default.svc.cluster.local
* kubectl top node node123
* kubectl top pod somepod --sort-by=cpu 
* kubectl top pod -A # all pods

Create a service for a pod of deployment:
* kubectl expose pod mypod --port 80
* kubectl expose deploy some-deployment --port 443

More complex debugging, interacting with etcd:
* from: https://github.com/k3s-io/k3s/issues/2732
* kubectl run --rm --tty --stdin --image docker.io/bitnami/etcd:latest etcdctl --overrides='{"apiVersion":"v1","kind":"Pod","spec":{"hostNetwork":true,"restartPolicy":"Never","securityContext":{"runAsUser":0,"runAsGroup":0},"containers":[{"command":["/bin/bash"],"image":"docker.io/bitnami/etcd:latest","name":"etcdctl","stdin":true,"stdinOnce":true,"tty":true,"volumeMounts":[{"mountPath":"/var/lib/rancher","name":"var-lib-rancher"}]}],"volumes":[{"name":"var-lib-rancher","hostPath":{"path":"/var/lib/rancher","type":"Directory"}}]}}'
* ./bin/etcdctl --key /var/lib/rancher/k3s/server/tls/etcd/client.key --cert /var/lib/rancher/k3s/server/tls/etcd/client.crt --cacert /var/lib/rancher/k3s/server/tls/etcd/server-ca.crt endpoint status

# Secrets

BEWARE: Prior to Kubernetes 1.7+, secrets are held in tempfs and under RBAC, but they are PLAINTEXT in etcd! Inspecting raw etcd can be tedious to track down depending on your k8s environment, but be aware of the tools and the threat model to etcd. 
* $ ETCDCTL_API=3 etcdctl --endpoints 127.0.0.1:2379 --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key --cacert=/etc/kubernetes/pki/etcd/ca.crt get /registry/secrets/default/default-token-t7j4c
* Consider etcdctl in your threat model.

Secrets are held in tempfs, and in pods at /var/run/secrets.
The default service account is mounted into pods by default, unless specified otherwise.
Service accounts possess three secrets:
* ca.crt: a cert by which to trust the api
* namespace: the namespace to which the service account belongs
* token: the authentication token

Secrets are stored in unencrypted base64 form.
In etcd, secrets are stored in plaintext! Although this may be subject to change.


Generating certs:
* openssl genrsa -out tls.key 2048
* openssl req -new -x509 -key tls.key -out tls.cert -days 365 -subj /CN=example.com  # CN is as mentioned in Ingress spec, and in general, is the name according to whomever the user is (an Istio component, k8s Ingress, etc)
* kubectl create secret tls ingress-cert --cert=tls.cert --key=tls.key

Generating generic






