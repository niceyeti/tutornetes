
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




