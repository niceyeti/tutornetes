#!/bin/bash

# Note: the cluster name created here corresponds with several scripts; grep when updating.
cluster_name="devcluster"
# Note: k3d prefixes many objects it creates 'k3d-', but should still be referred via their unprefixed name.
prefixed_cluster_name="k3d-$cluster_name"

function cleanup() {
    echo "Deleting ALL clusters."
    clusters=$(k3d cluster list -o json | jq -r .[].name)
    echo "Deleting $clusters"
    k3d cluster delete $clusters
    cat <<- EOF
Cluster deleted.
Note: although the cluster was deleted, you may want to manually prune images, volumes etc with:
* docker system prune -a
* docker image prune -a
But be sure you know what these commands do before executing, or you may accidentally delete desired
containers and images, such as your dev images/containers. Likewise, you'll have to re-pull many base
k3d images.
EOF
}

function bail_if_cluster_exists() {
    clusters=$(k3d cluster list -o json | jq -r .[].name)
    for cluster in $clusters; do
        if [[ $cluster == $cluster_name ]]; then
            echo "Cluster >$cluster_name< already exists. Delete it before creating."
            exit
        fi
    done
}

function check_prereqs() {
    if [[ -z $(which helm) ]]; then 
        echo "Helm must be installed (for istio, et al), exiting..."
        exit
    fi
}

# Creates the cluster and all extended properties (helm charts, istio, etc.).
# The latter could be broken out, but this is fine for now.
function create() {
    check_prereqs
    bail_if_cluster_exists

    echo "Creating development cluster '$cluster_name'..."
    # Create a k3d cluster with n worker nodes. See: https://www.suse.com/c/introduction-k3d-run-k3s-docker-src/
    # "k3d waits until everything is ready, pulls the Kubeconfig from the cluster and merges it with your default Kubeconfig"
    # Note: all ports exposed on the serverlb ("loadbalancer") will be proxied to the same ports on all server nodes in the cluster
    #k3d cluster create --config k3d_config.yaml
    k3d cluster create --config k3d_config.yaml
    echo "Cluster created."
    echo ""
    install_helm_charts
    cluster_info
}

function cluster_info() {
    echo ""
    echo "HELM CHARTS:"
    helm list --all-namespaces
    echo ""

    echo ""
    echo "CRDS: "
    kubectl get crds --all-namespaces
    echo ""

    echo ""
    echo "PODS: "
    kubectl get pods --all-namespaces
    echo ""

    echo "WORKFLOW NOTE: when leaving/restoring the dev environment, avoid re-pulling images by using the '--pause' and '--restart' flags."
}

function install_helm_charts() {
    echo "Installing helm charts..."

    helm repo add istio https://istio-release.storage.googleapis.com/charts
    helm repo update

    kubectl create namespace dev
    # all apps deployed to dev will have Envoy
    kubectl label namespace dev istio-injection=enabled

    # install the istio base chart, which is most of its control plane components.
    # presumably this will also add all of the istio CRDs (virtualservice, gateway, etc.)
    kubectl create namespace istio-system
    helm install istio-base istio/base -n istio-system
    # install the istio discovery chart
    helm install istiod istio/istiod -n istio-system --wait
    # install ingress gateway
    kubectl create namespace istio-ingress
    kubectl label namespace istio-ingress istio-injection=enabled
    helm install istio-ingress istio/gateway -n istio-ingress --wait

    # install prometheus and kiali; this is fragile and unsecure, so simply disable lines if not needed.
    # I'm only adding it here by default since I'm actively playing with istio.
    echo "Installing kiali and prometheus..."
    kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.13/samples/addons/kiali.yaml
    kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.13/samples/addons/prometheus.yaml
    # To delete kiali and prometheus, simply use the delete verb per the above:
    # kubectl delete -f https://raw.githubusercontent.com/istio/istio/release-1.13/samples/addons/kiali.yaml
    # kubectl delete -f https://raw.githubusercontent.com/istio/istio/release-1.13/samples/addons/prometheus.yaml
}

function pause() {
    echo "Stopping $prefixed_cluster_name cluster."
    k3d cluster stop $cluster_name
    echo "Pause completed."
    echo "Use './up.sh --restart' to restart."
}

function restart() {
    k3d cluster start $cluster_name
    echo "Cluster $prefixed_cluster_name restarted."
}

function show_help() {
cat <<- EOF
Commands:
    * Create a cluster from scratch:
        ./up.sh --new
    * Restart a cluster:
        ./up.sh --restart
    * Delete ALL clusters, including k3d-managed registries:
        ./up.sh --clean
    * Pause the cluster, saving the registry as well:
        ./up.sh --pause
    * Updating config/k3d:
        Updating is best done manually, to refresh the fundamental components.
        Use `k3d config migrate k3d_config.yaml new_config.yaml`

Workflow: it is best to rely on the k3d api to manage all cluster resources, instead of implementing
scripts to do so, especially as the k3d api evolves. From-scratch cluster+registry creation is problematic
since the bare registry has to re-pull all base k3s and k3d images each time you create a new
cluster and registry.

The basic KISS usage of k3d is best: run create to make a cluster and registry, then run 'stop'
and 'start' to save or restore the cluster. This will retain the registry. IOW, let k3d manage the cluster
and registry together, don't manage the dev registry separately---let the tool do the work.
EOF
}

for arg in "$@"
do  
    if [[ $arg == "--new" || $arg == "--create" ]]; then
        create
        exit
    fi

    if [[ $arg == "--clean" || $arg == "--delete" ]]; then
        cleanup
        exit
    fi

    if [[ $arg == "--pause" || $arg == "--stop" ]]; then
        pause
        exit
    fi

    if [[ $arg == "--restart" || $arg == "--start" ]]; then
        restart
        exit
    fi

    if [[ $arg == "--help" || $arg == "-h" || $arg == "help" ]]; then
        show_help
        exit
    fi
done

show_help