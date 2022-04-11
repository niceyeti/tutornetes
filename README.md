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

## Soapbox

Probably the most difficult part of learning Kubernetes and other CNCF technology is learning and configuring the many resources required to do so, while navigating materials polluted by self-promotion and pseudo-professional linkedin clickbait. Many organizations simply don't want to learn, nor invest the time/resources required, but instead drive up their technical debt via a code-first-ask-questions-later attitude.

This repo provides a template for developing k8s clusters and cloud applications using k3d, helm, and tilt.
The objective is that you can branch off the base_cluster project, modify it to your deployment/app/infrastructure/etc,
and rapidly develop new clusters, charts, and so forth. So for example, create a branch, run the base_cluster
to ensure your environment is consistent, then start modifying the base_cluster to rapidly prototype a new cluster, app, 
chart, etc. The repo includes builtin chart/cluster scanning with kubescape to provide security linting.

This repo's goals are pure research and development:
1) the ability to design, develop, and spin-up clusters with different properties
2) to rapidly design, develop and test cloud apps themselves
3) to provide a learning environment, your own personal k8s playground
4) minimal free-climbing: pushing quality concerns as far upstream in the development process as possible, with immediate development feedback and security/quality scanning.

## Repo Organization
* */clusters*: create folders here containing content (helm charts, k3d startup, etc). describing a cluster
    * */base_cluster*: an example cluster
        * */go_app*: an example golang cloud application with some basic endpoints and a static page
        * */scripts*: script cemetery for random or obsolete scripts

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
Install docker, k3d, helm, tilt, kubescape (optional), and istio (if used). See version_info.txt for versions.

## Basic Workflow
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
I set this up to be run manually because it is easiest to maintain and I am unlikely to keep updated; a fully-fledged CI system would run the scanner as a formal part of the linting process, like any linter or test-success exit conditions. Note that kubescape is not an offline tool; see its docs to make sure you understand its remote interactions and reporting.
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
    * Updating: same as installation, just make review and update how the istio manifest is committed; the intent is simply to ensure that the istio version is in the repo, and no other istio artifacts.

## DevOps Resources
When learning kubernetes I consciously avoided online materials entirely and focused solely on books. Some I found most useful:
1) [Kubernetes In Action](https://www.amazon.com/Kubernetes-Action-Marko-Luksa/dp/1617293725/)
2) [Kubernetes Patterns](https://www.amazon.com/Kubernetes-Patterns-Designing-Cloud-Native-Applications/dp/1492050288/)
3) [Design Patterns for Container-Based Distributed Systems](https://www.usenix.org/conference/hotcloud16/workshop-program/presentation/burns) (free and a quick read)
4) [K3d](https://k3d.io/v5.1.0/): [quick tutorial](https://www.youtube.com/watch?v=mCesuGk-Fks)
5) [Helm](https://helm.sh/docs/intro/quickstart/): [quick tutorial](https://www.youtube.com/watch?v=5_J7RWLLVeQ)
6) [Tilt](https://tilt.dev/): [simple k3d tilt example](https://github.com/iwilltry42/k3d-demo/blob/main/Tiltfile)
7) [Istio Up and Running](https://www.amazon.com/Istio-Running-Service-Connect-Control/dp/1492043788/)

## Credit
This repo was gratefully built atop k3d, docker, tilt, helm, k3s, kubescape, and kubernetes--and google as well. All credit for these tools goes to their authors. Seriously, thanks a ton! Those who teach instead of tell deserve utmost praise.

Some very helpful teachers:
* https://www.youtube.com/c/MarcelDempers
* https://www.youtube.com/c/DevOpsToolkit
