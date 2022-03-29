# Kube Operators

## Intro
Creating CRDs and the Operators that act upon them is complex and api-dependent; in other words it is subject to change. These notes capture the basic approach using the Operator SDK. This is primarily based off of examples given in "Programming Kubernetes" by Hausenblaas and Schimanski. I have explicitly plagiarized and updated components of the authors' CRDs/Operator/code-generation. This example follows the authors' example of the following CRD/Operator definition:
* An "At" CRD defining a command to run and a time to run it.
* An operator implementing the business logic that occurs when an At changes.
Credit goes to the authors', adapted under Apache 2.0, whose repo is here: https://github.com/programming-kubernetes

Note that Operators are not the only "plugins" for kubernetes. There are plugins for the kubelet, kube proxy, and so forth.

## Version Boilerplate
The implementation rests largely on code generation, which though volatile in terms of learning (who knows if code generation will even be necessary after golang templates...) is still a worthwhile exercise to understanding the kubernetes core components and their interactions: etcd, the api server, CRDs, operators, etc. The book's operator-sdk usage is out of date; this is an attempt to update it. The operator-sdk project has legacy docs and other info w.r.t. updating commands from the 0.x version used in the book: 
* https://v0-19-x.sdk.operatorframework.io/docs/golang/legacy/quickstart/
* https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/
* google any errors, and others' feedback has probably addressed any missing/updated commands


## Requirements
1) Install operator-sdk. I installed via compiled exe via these steps: https://sdk.operatorframework.io/docs/installation/
2) See operator-sdk golang and kubernetes/kubectl reqs.

## Creating an 'At' operator
Note that these instructions will easily become out of date. I also bootstrapped them with a lot of 'this seems to work' hammer-it-til-it-runs preparation; the command line options should not be treated as complete. 
1) Create the CRD itself:
* `kubectl apply -f crds/cnat_v1_alpha1_at_crd.yaml`
2) With the definition declared, create an "At" resource:
* `kubectl apply -f crds/cnat_v1_alpha1_at_crd.yaml`
3) Initialize the operator project. This runs code generators to generate a big pile of boilerplate and other resources, so don't be surprised:
* `operator-sdk init cnat-operator --domain example.com --repo github.com/niceyeti/tutornetes --license apache2`
4) Create the api and controller. Originally this was the two commands `add api` and `add controller` on p. 124 of Programming Kubernetes:
* `operator-sdk create api --version v1 --kind At --group cnat.programming-kubernetes.info --resource --controller`

# Left off here. The following steps are tentative, still figuring them out with lots of freeclimbing.
5') Build go build .

Hmm... port conflicts 



Left off here. TODO:
adapt this command
operator-sdk add controller --api-version ... p. 124
Compare with:
https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/

Tentative:
- Build the controller
- Run it outside the cluster using any different metrics address than the docker proxy (0.0.0.0:8080)
  ./tutornetes -metrics-bind-address 0.0.0.0:8082
- address errors

TODO:
0) Code for controller and go types
0) dockerize and ship the controller
1) Fully understand/characterize api objects and their relationships, at least per controllers.
2) Dependency mgt and updating














