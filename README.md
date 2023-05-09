## About

The difficult part of learning Kubernetes and CNCF components is learning and configuring the many resources required. You dip in a toe only to get swept away by the tide of components, languages, and development tools! On the other hand, learning 'all' of DevSecOps is such a rapidly moving target that a flexible dev/learning environment and a positive attitude toward learning carry one further than any certification.

This repo provides a template for developing k8s clusters and cloud applications using k3d, istio, helm, and tilt. The base environment allows you to run a complete k8s cluster locally under k3d, with live updates using tilt, and a bunch of devsecops bells and whistles built into the devcontainer and tilt artifacts for cluster security scanning with kubescape.

The objective is that you can branch off the base_cluster project, modify it to your deployment/app/infrastructure/etc, and rapidly develop new clusters, charts, and so forth, prototype a bit and then create an independent repo containing only what you need.

So for example, create a branch, run the base_cluster
to ensure your environment is consistent, then start modifying the base_cluster to rapidly prototype a new cluster, app, chart, etc. The repo includes builtin chart/cluster scanning with kubescape to provide security linting.

This repo's goals are pure dev research and training:

1) the ability to design, develop, and spin-up clusters with different properties
2) to rapidly design, develop and test cloud apps themselves
3) refine and tailor the devsecops artifacts that you need
4) to provide a learning environment, your own personal k8s playground
5) minimal free-climbing: pushing quality concerns as far upstream in the development process as possible, with immediate development feedback and security/quality scanning.

## Repo Organization
* */clusters*: create folders here containing content (helm charts, k3d startup, etc). describing a cluster
    * */base_cluster*: an example cluster
        * */go_app*: an example golang cloud application with some basic endpoints and a static page
        * */go_grpc_example*: and example gRPC CRUD application in golang and backed by postgres. A good integration test example using dockertest is included.
        * */go_webhook_example*: A minimal working example of a kubernetes admission webhook, in golang. The webhook merely logs a message, but could amend the object definitions however one needs.
        * */tools_container*: A handy container and Pod definition based on the tutum/dnsutils containers, useful for debugging cluster components. Any other one-off diagnostic tools one might need can be added, built, and deployed.
        * */scripts*: script cemetery for random or obsolete scripts
        * */ops_extras*: Not yet fully defined, not a first-class project in the repo. Currently I use this to store declarative examples of role-bindings, etc.
    * */misc*: an unhealthy collection of review notes and other stuff exposing my own ignorance :)
The primary resources to understand are in the *base_cluster* folder:
| Resource | Description |
| :--- | :--- |
| *k3d_config.yaml* | The base cluster definition. Modify this to create clusters with different properties: nodes, volumes, etc. |
| *up.sh* | Defines starting, stopping, or deleting the base cluster |
| *Tiltfile* | Drives tilt. Apps and development resources are defined here. |
| *go_app* |  Contains the helm chart and source code for an example app   |
| *go_app/chart* |  The helm chart and params for creating the go app's k8s objects. |
| *go_app/src* |  The source code for an extremely simple golang webapp; basically a dockerfile and a few dummy endpoints. |

## Pre-reqs
The development container provides full golang development support, just open the tutornetes/ folder in vscode and build the dev container.

However these components are required on your host (outside the dev container):
* docker, k3d, helm, tilt, kubescape (optional), and istio (if used). 
* See version_info.txt for versions.

I have not specified all reqs in the build container, primarily because I don't want to run docker in the dev container.
As such, some things are intended to be run from a host satisfying those reqs, where noted.

## Basic Workflow

Note: these steps are performed on the host machine, as I haven't fully containerized the dev environment, and most likely won't do so due to its size and volatile dependencies.

1) cd into *base_cluster*
2) create a k3d-development cluster: `./up.sh --create`
3) wait for cluster creation to complete: `kubectl get pods --all-namespaces`. It may take a few minutes to initially pull and install the cluster images (k3s, traefik, etc); pods will be shown as "ContainerCreating"; the cluster is ready when all pods show "Running" or "Completed".
4) poke around to get an idea of the cluster components:
    * view the nodes and registry: `docker container ls --all`
    * view the cluster running inside the nodes: `kubectl get pods -o wide --all-namespaces`
    * view host interfaces: `ifconfig`
    * know the namespaces (otherwise they will trip you): `kubectl get namespaces`
5) tilt up to run the go app:
    * `tilt up`, then open a browser
    * navigate to `localhost:10350`
    * after a minute or so, hit the go app at `localhost:8080/fortune`
6) Optional: if kubescape is installed, use it to get a report on the security posture of the cluster.
I set this up to be run manually because it is easiest to maintain and I am unlikely to keep updated; a fully-fledged CI system would run the scanner as a formal part of the linting process, like any linter or test-success exit conditions. Note that kubescape is not an offline tool, and shares your info with third party; see its docs to make sure you understand its remote interactions and reporting.
    * navigate to `localhost:10350`
    * click to run the cluster scan; view the results and behold all of the things you have to spend the next week fixing! :P
    * click to run the app scan; this method scans only the go app and its chart, more relevant to the developer than the entire cluster
    * See: https://kubernetes.io/blog/2021/10/05/nsa-cisa-kubernetes-hardening-guidance/


## Updating and Maintenance
* Updating k3d: after updating k3d and k3s, run `k3d config migrate k3d_config.yaml new_config.yaml` and review the new config, then commit it. The migration code itself can be reviewed in the k3d repo.
* Installing and updating istio: istio was installed using the installation directions [here](https://istio.io/latest/docs/setup/getting-started/). I installed it to the misc/ directory so the version is part of the repository, not the system. The istio folder contains many examples: web app, web sockets, operator, external access, and more.
    * Installation:
        * cd misc
        * curl -L https://istio.io/downloadIstio | sh -
        * cd istio-[version]/
        * export PATH=$PATH:$pwd/bin
    * Updating: same as installation, just review and update how the istio manifest is committed; the intent is simply to ensure that the istio version is in the repo, and no other istio artifacts.

## DevOps Resources
When learning kubernetes I consciously avoided online materials entirely and focused solely on books. Some I found most useful:
1) [Kubernetes In Action](https://www.amazon.com/Kubernetes-Action-Marko-Luksa/dp/1617293725/)
2) [Kubernetes Patterns](https://www.amazon.com/Kubernetes-Patterns-Designing-Cloud-Native-Applications/dp/1492050288/)
3) [Design Patterns for Container-Based Distributed Systems](https://www.usenix.org/conference/hotcloud16/workshop-program/presentation/burns) (free and a quick read)
4) [K3d](https://k3d.io/v5.1.0/): [quick tutorial](https://www.youtube.com/watch?v=mCesuGk-Fks)
5) [Helm](https://helm.sh/docs/intro/quickstart/): [quick tutorial](https://www.youtube.com/watch?v=5_J7RWLLVeQ)
6) [Tilt](https://tilt.dev/): [simple k3d tilt example](https://github.com/iwilltry42/k3d-demo/blob/main/Tiltfile)
7) [Istio Up and Running](https://www.amazon.com/Istio-Running-Service-Connect-Control/dp/1492043788/)
8) [Cloud Native Go](https://www.amazon.com/Cloud-Native-Go-Unreliable-Environments/dp/1492076333) (A Golang-specific book, but nonetheless a terrific resource on general architecture and problems)


## Credit
This repo was gratefully built atop k3d, docker, tilt, helm, k3s, kubescape, istio, and kubernetes, lots of hard work by Google, Mirantis, ArmoSec, Tilt, and others.
Any copyright/license issues (cloud-native source licenses are prone to 'upgrade') are unintentional, and this repo is for non-commercial use.
All credit for these technologies goes to their authors, with sincere thanks. We stand kittens on the shoulders of giants and call ourselves lions, lol.

Those who teach instead of tell deserve utmost praise. Some helpful teachers:
* https://www.youtube.com/c/MarcelDempers
* https://www.youtube.com/c/DevOpsToolkit

![image](wrench.png)

```
     ____        _ ____     _   __      __     ____                    __    __  
    / __ )__  __(_) / /_   / | / /___  / /_   / __ )____  __  ______ _/ /_  / /_ 
   / __  / / / / / / __/  /  |/ / __ \/ __/  / __  / __ \/ / / / __ `/ __ \/ __/ 
  / /_/ / /_/ / / / /_   / /|  / /_/ / /_   / /_/ / /_/ / /_/ / /_/ / / / / /_   
 /_____/\__,_/_/_/\__/  /_/ |_/\____/\__/  /_____/\____/\__,_/\__, /_/ /_/\__/   
                                                            /____/               
```
... and frequently 'borrowed'.
