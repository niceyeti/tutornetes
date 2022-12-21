# goop

Goop is a simple golang operator, purely for exercise, and the value of the business logic is basically nil.
This operator merely simulates the management of a stateful application,
by deploying a few dummy Job objects when a Goop object is created:
1) Goop object is created
2) Goop controller observes creation and deploys Goop.JobCount 'worker' Pods.
These are merely busybox 'worker' containers to which 'cmd' strings are passed and printed.
3) Goop controller observes completion of all the Jobs and marks the Goop object as "Completed"

This is a simple example of a stateful object, needlessly complicated by operator code
that could be replaced by k8s primitives instead.
But it can be easily extended to more complex patterns, such 
as implementing various forms of Job logic and dependencies,
such as for node configuration, db migration, or a build system.
Kubernetes has no native distributed-job object for ensuring that node-based jobs are distributed
across nodes, but a modicum of utility of this operator is that it achieves this using daemonsets.

## Current State (delete when controller complete)
The following commands run the controller in the cluster, but there are missing resources/roles:
- make deploy        # make manifest
- make docker-build  # make the controller image and push it
- make docker-push   # make sure the image is available in the cluster
- make deploy # TODO: not sure the order or deploy vs install; check the makefile
- push any other docker images (busybox, pause, etc.)
- make install       # installs CRDs and runs the controller
- kubectl create -f temp/role.yaml
- kubectl create -f temp/test_goop.yaml

The issue is that the controller is currently:
1) missing cluster roles for daemonsets. I added a role to temp/role.yaml but this should be
integrated into the native manifest generation somehow.
2) the deployment has the image as "controller:latest" but needs registry prefix for my cluster: "k3d-devregistry:5000/controller:latest"
    * FIX: in config/manager/manager.yaml, simply add the required prefix to the image field.
Both of these should be resolved such that no manual steps are required to run the controller.
I think this may have to do with kustomize?

## Dev notes
- I had to modify the image name in config/manager/manager.yaml to have the k3d-devregistry:5000 image prefix
- I had to create clusterroles to allow the controller to query and create daemonsets: see temp/role.yaml
    * do not namespace clusterroles, nor clusterrolebindings
    * daemonset api group is 'apps'




## Goal

Recall the actual logic of controllers can be seen as merely automated kubectl'ing.
The goal is to understand the layers and patterns by which a controller could be abstracted
to perform arbitrarily complex operations. A useful way to look at the value of controllers
is that they could be used to implement distributed linux commands; while most Operators chase
business value, it is important to think about how they could be used for system value, e.g. 
for recurring system maintainenance tasks, a distributed cmd line.

## Do you need an Operator?

Note: many Job patterns have already been considered and developed in great detail with off-the-shelf patterns.
The redis work-queue example could be easily adapted to implement distributed
Jobs, such as web crawlers attached to Nodes with independent network interfaces for better throughput or rate-limit diversification.
See:
* https://kubernetes.io/docs/tasks/job/.
* https://kubernetes.io/docs/concepts/workloads/controllers/job/#job-patterns
* https://github.com/kubernetes/kubernetes/issues/36601

In most cases, one should develop one's needs using kubectl and native k8s
objects. Then, where k8s primitives are not supported, change one's requirements/features/use-cases
until supported, if possible. Based on this composition of commands+objects, only then
implement an Operator--if it would still even be necessary! Code is vastly more heavyweight
and less reusable than declarative commands+objects; the goal is always to mercilessly minimize
or eliminate code wherever a declarative patterns exists.

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
    * most the code is straightforward copy/paste from available client examples, to simulate
    kubectl commands. The difficulty are unusual state update loops and error cases, which may
    be discovered simply by running the controller--however, expect such integration/state issues
    to occur, and code defensively by using idempotent ops and avoiding unusual or convoluted
    logic.

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

Operator-SDK merely uses kubebuilder under the hood. Despite being a bit overwhelming at first glance, in terms
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
* a high quality hand-holding expedition: https://pres.metamagical.dev/kubecon-eu-2019.pdf

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

