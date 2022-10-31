# Intro 
The simple_app folder is a demo learning example for developing a kubernetes based golang http app using helm, tilt, and k3d. The kubernetes objects for the app are extremely simple, the intent is for this to serve as a starting point for developing more complex apps and to write/develop/test specific cluster properties, charts, and kubes features.

## TODO
1) Add secrets/https
2) Add postgres chart
3) Tracing and logging... in the well-worn tradition of under-investing in critical
components, I probably won't add this in order to keep this repo barebones KISS.

## Steps to Run
1) Start the k3d cluster and registry:
    * `./up.sh`
2) Run tilt to start the watch/build/deploy loop:
    * `tilt up`
    * ctrl+C and then `tilt down` to remove the old artifacts

## Other
* Teardown and clean up: `./up.sh --clean`
* Run a specific app: `tilt up simple-go-app`
* Run any chart the lazy way, e.g. postgres: `helm install my_pg bitnami/postgresql`

## Notes

Helm:
* To verify chart, generate its yamls: `helm template [chart name] [chart directory]`
    * `helm template simple_app .`
    * `helm lint simple_app .`
    * `helm template . --debug --values Values.yaml`
* To view some chart, such as a bitnami:
    * `helm show chart bitnami/postgresql`

Registries:
This is complicated because of the pointers to registries in the charts, and the credential handling involved.
Registry and cluster names should be grepped since they often exist across yaml files and scripts.
* See https://k3d.io/v5.0.3/usage/registries/
* The simplest and best usage of a registry is to let k3d manage it, by using `cluster start/stop` instead of writing scripts to manage some registry.
* Run:
    1) Run a private registry in the cluster:
    ```
    # Create a cluster and a registry
    k3d cluster create mycluster --registry-create mycluster-registry
    # Find the port on which the registry is serving
    docker ps -f name=mycluster-registry
    ```
    2) Pull an image and retag it using the image name path prefix convention to tell containerd where to get it.
    This could also just be a locally-built image, but for the sake of example:
    ```
    # some image
    docker pull alpine:latest
    # re-tag it with the registry from (1) as its prefix
    docker tag alpine:latest mycluster-registry:12345/testimage:local
    # push it to the registry in the cluster
    docker push mycluster-registry:12345/testimage:local
    # test that the image can be run in the cluster
    kubectl run --image mycluster-registry:12345/testimage:local testimage --command -- tail -f /dev/null
    ```

Tilt:
    # docker build -t companyname/frontend ./frontend
    docker_build("companyname/frontend", "frontend")

    # docker build -t companyname/frontend -f frontend/Dockerfile.dev frontend
    docker_build("companyname/frontend", "frontend", dockerfile="frontend/Dockerfile.dev")

    # docker build -t companyname/frontend --build-arg target=local frontend
    docker_build("companyname/frontend", "frontend", build_args={"target": "local"})







