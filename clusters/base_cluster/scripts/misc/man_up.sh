#!/bin/bash

# This is the procedural method for starting k3d, creating a registry, building docker app image, and manually deploying
# using kubectl commands. This was used to develop the needed objects and build steps, then gradually migrate these to 
# Tilt/Helm and declarative resources. It is still useful for testing the dev environment.

cluster_name="mycluster"

function cleanup() {
    echo "Cleaning up..."
    clusters=$(k3d cluster list -o json | jq -r .[].name)
    echo "Deleting $clusters"
    k3d cluster delete $clusters
}

cleanup

for arg in "$@"
do
    if [ $arg == "--clean" ]; then
        echo "Clean completed"
        exit
    fi
done


# TODO: somehow parameterize the registry location; overall cleanup the registry semantics here (move out of here somehow? run docker in cluster, mount in code, run there? naw, not very tilty...)
# Cleanup probably relies on how Tilt wants/runs a registry; they may have something baked in.
cd code
docker build -f Dockerfile . -t simple-go-app
cd ..

# https://k3d.io/usage/guides/exposing_services/


# Create a k3d cluster with n worker nodes. See: https://www.suse.com/c/introduction-k3d-run-k3s-docker-src/
# "k3d waits until everything is ready, pulls the Kubeconfig from the cluster and merges it with your default Kubeconfig"
# Note that the --registry-create command is supposed to accept a name parameter, but is currently broken in my k3d version
# FUTURE: understand and implement the --cluster-init switch to swap between etcd or sqlite; currently using defaults. HA would make a good project.
#k3d cluster create $cluster_name --agents 1 -p 8080:30080@agent[0] --registry-create --wait
k3d cluster create $cluster_name --api-port 6550 -p "8081:80@loadbalancer" --agents 1 --registry-create --wait  # see  https://k3d.io/usage/guides/exposing_services/
# Note: all ports exposed on the serverlb ("loadbalancer") will be proxied to the same ports on all server nodes in the cluster
# Import images into the cluster the procedural way; TODO: I think this is redundant with the --registry-create above
#exit
#docker ps -f name=k3d-mycluster-registry # get the port (42783 below)
registry_host_port=$(docker inspect --format '{{(index (index .HostConfig.PortBindings "5000/tcp") 0).HostPort}}' $(docker ps -q -f name=k3d-mycluster-registry))
cluster_image_name="k3d-$cluster_name-registry:$registry_host_port/simple-go-app"
echo "Attempting to load app image into cluster registry: $cluster_image_name"
docker tag simple-go-app $cluster_image_name
#k3d image import $cluster_image_name -c $cluster_name
docker push $cluster_image_name
kubectl create deployment simple-go-app --image=$cluster_image_name
kubectl create service clusterip simple-go-app --tcp=80:80
kubectl create -f ingress.yaml
# Test an endpoint in the app.
# NOTE: sleeping is required because my app is not yet implementing deterministic up-checks. Some object is taking
# about 20s to catch up before the app begins serving. This is understandable because the raw kube commands above
# do not arrange health/liveness checks, etc.
echo "Sleeping for 20s because my kube objects are not behaving deterministically, are catching up to consistency, or my app is not implementing health checks properly..."
sleep 20s
curl http://localhost:8081/fortune
exit

# This method works:
# Prelim:
#    k3d cluster create $cluster_name --agents 1 -p 8080:30080@agent[0] --registry-create --wait
#    docker tag simple-go-app k3d-mycluster-registry:42783/simple-go-app
#    k3d image import k3d-mycluster-registry:42783/simple-go-app -c mycluster
# 1) edit /etc/hosts to redirect k3d-mycluster-registry to 127.0.0.1
# 2) docker tag simple-go-app k3d-mycluster-registry:42783/simple-go-app
# 3) docker push k3d-mycluster-registry:42783/simple-go-app
# 4) kubectl run simple-go-app --image=k3d-mycluster-registry:42783/simple-go-app -n default
# 5) kubectl expose pod simple-go-app --port=30080 --target-port=8080
# 6) kubectl get pods -n default --watch
# To then reach the pod, 

# Two problems to solve: 1) cleanup registry and my understanding of it 2) connectivity to app in cluster from host

# Debugging: the good ole pod piggyback: kubectl run nettools --image=someimage -- curl simple-go-app/fortune
# Exec into docker container agent 'node':
#     kubectl descrie pod simple-go-app  <-- get the cluster ip of the app pod
#     docker exec -it bec3 /bin/sh       <-- exec inside the cluster
#     wget -qO- 10.42.1.11:8080/fortune  <-- test an endpoint served by the app


# kubectl run simple-go-app --image=simple-go-app:latest -n default
# kubectl expose pod simple-go-app --port=30080 --target-port=8080



# Options:
# --api-port 127.0.0.1:6445  map the kubes server api port (6443 internally) to 6445 on localhost; kubeconfig will contain connection string: https://127.0.0.1:6445
# --volume /home/me/mycode:/code@agent[*]  bind mount /home/me/mycode into /code on agents, e.g. for hot-reload app development
# --port '8080:80@loadbalancer' maps localhost’s port 8080 to port 80 on the load balancer (serverlb), which can be used to forward HTTP ingress traffic to your cluster. you can now deploy a web app into the cluster (Deployment), which is exposed (Service) externally via an Ingress such as 'myapp.k3d.localhost'.
# Then (provided that everything is set up to resolve that domain to your local host IP), you can point your browser to http://myapp.k3d.localhost:8080 to access your app. Traffic then flows from your host through the Docker bridge interface to the load balancer. From there, it’s proxied to the cluster, where it passes via Ingress and Service to your application Pod.


# Get the exposed port of the registry (probably 5000? can this be specfied in the previous command?)
docker ps -f name=mycluster-registry
exit

# NOTE: get the correct port from the above ps command. 32863 is a placeholder below.
# TODO: I think this is the sequence that worked, assuming you've built simple-go-app image:
docker tag simple-go-app localhost:32863/simple-go-app
docker push localhost:32863/simple-go-app  # important: push to the exposed port on localhost; the image itself contains the correct tag for resolving inside the cluster
kubectl run simple-go-app --image k3d-mycluster-registry:5000/simple-go-app -n default # Note the image is pushed to the registry on localhost, but run via the registry hostname inside the cluster
#kubectl get pods --all-namespaces --watch
kubectl expose pod simple-go-app --port=30080 --target-port=8080 #expose the pod; note this is somewhat incorrect, you'll want to expose, say, the deployment; this is just hacking.
# The above is close, but no cigar. https://stackoverflow.com/questions/68547804/how-to-expose-two-apps-services-over-unique-ports-with-k3d
# "Without opening a port during the creation of the k3d cluster, a nodeport service will not expose your app"
# The k3d docs were pretty clear on this topic, I just need to review them.



# Push the image to the registry in the cluster
# TODO: this is incorrect/placeholder. How to get the port?
docker push registry:5000/simple-go-app
exit

# Show the template output; this will also show errors, if any
helm template simple_app .
exit

# Install the app. Passing the values file is just a self-reminder that you can define multiple values files...
helm install simple_app . -f values.yaml 

# Not needed, but to upgrade a running app:
# helm upgrade simple_app . --values some-other-values.yaml















