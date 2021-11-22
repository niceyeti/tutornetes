#!/bin/bash
set -eux

# Remove all containers named "simple-go-app"
function cleanup() {
    echo "cleaning up..."
    container_ids=$(docker container ls -aq --filter "name=simple-go-app")
    docker container rm --force $container_ids
    echo "done."
}

trap cleanup SIGINT

./build.sh && docker run --name simple-go-app -p 127.0.0.1:80:8080/tcp simple-go-app
