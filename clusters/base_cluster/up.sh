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

function create() {

    bail_if_cluster_exists

    echo "Creating cluster '$cluster_name'..."
    # Create a k3d cluster with n worker nodes. See: https://www.suse.com/c/introduction-k3d-run-k3s-docker-src/
    # "k3d waits until everything is ready, pulls the Kubeconfig from the cluster and merges it with your default Kubeconfig"
    # Note: all ports exposed on the serverlb ("loadbalancer") will be proxied to the same ports on all server nodes in the cluster

    echo "Creating cluster..."
    k3d cluster create --config k3d_config.yaml
    echo "Cluster created."
    echo ""
    echo "WORKFLOW NOTE: when leaving/bring-back the dev environment, avoid re-pulling images by using '--pause' and '--restart' flags."
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