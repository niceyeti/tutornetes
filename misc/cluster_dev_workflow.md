# Cluster Dev Workflow

The goal of the base_cluster is to define a vanilla cluster implementing the full stack
of development components:
1) fixed cluster and infrastructure components: up script, istio, helming
2) devsecops basic resources: tilt, kubescape
3) vanilla web app: a golang app with some endpoints, basic root page, and its charts
4) tools: the debugging container

To develop a complete stack, one branches from master to introduce a new app; any changes
to infrastructure should be merged back to master as needed to keep it up to date.
This imperative approach is fine, since I'm working solo.





