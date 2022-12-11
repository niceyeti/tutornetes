# goop

Goop is a simple golang operator, for exercise.
This operator merely simulates the management of a stateful application,
by deploying a few dummy Job objects when a Goop object is created:

1) Goop object is created
2) Goop controller observes creation and deploys some Job objects (mere sleep commands in busybox containers)
3) Goop controller observes completion of the Jobs and marks the Goop object as "Completed"

This is an extremely simple example of a stateful object.
But it can be easily extended to more complex patterns, such 
as implementing various forms of Job logic and dependencies,
such as for node configuration, db migration, or a build system.

## Construction commands

Reference: a short, similar version is at https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/

1) operator-sdk init --domain example.com --repo github.com/example/goop
2) go work use .
3) operator-sdk create api --group goop --version v1alpha1 --kind Goop --resource --controller
4) implement api:
    - modify api/v1alpha1/goop_types
    - run `make generate`
    - once satisfied with types, run `make manifests` to generate CRDs
        * result: CRD is generated and written to config/crd/bases
    - run `go mod tidy` (this can be re-run as new code is generated)
5) implement the controller: 
    * reference implementation: https://github.com/operator-framework/operator-sdk/blob/latest/testdata/go/v3/memcached-operator/controllers/memcached_controller.go


# Custom Operator
This is a custom operator example described here: https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/

There are a few order of operations requirements and gotchas with environmental issues:
* init the operator, then add it as a go workspace. This must be done because init'ing the operator
  creates the go.mod files used by `go work use .`.
  * If not performed, attempting to generate the api will fail with complaints about the go.work file.
  * the complete order of commands (note also the folder name must match the go project, 'memcached-operator'):
    1) operator-sdk init --domain example.com --repo github.com/example/memcached-operator
    2) go work use .
    3) operator-sdk create api --group cache --version v1alpha1 --kind Memcached --resource --controller
    4) make docker-build docker-push IMG="127.0.0.1:5000/memcached-operator:v0.0.1"
        * NOTE: 127.0.0.1:5000 in this command is currently how one reaches the image repo running in the base_cluster

## Details
Operator-SDK merely uses kubebuilder under the hood. Despite being a bit overwhelming at first glance in terms
of scripts and code-generation workflows (which seem to change constantly), Operator-SDK projects are actually
somewhat easy to understand. Kubebuilder provides a more internal view of Operators and their programming requirements.

## Resources
    
Source for this project:
* https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/
* https://sdk.operatorframework.io/docs/
Apis, Groups, and Versioning are core technical concepts:
* https://book.kubebuilder.io/cronjob-tutorial/gvks.html
Kubebuilder is a great programmer resource for the internal guts of Operators.
* https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html





## Description
// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/goop:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/goop:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

