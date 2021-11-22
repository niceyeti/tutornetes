#!/bin/bash
echo "Pruning unused docker objects: volumes, networks, containers, images..."
docker system prune --volumes
echo ""
echo "Also consider running 'docker image prune -a' manually, after verifying no desired images will be deleted."
echo ""