
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
k create pod my-pod --image=busybox --restart=Never --command 



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

