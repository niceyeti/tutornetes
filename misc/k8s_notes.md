An extended cheatsheet of kubernetes commands and objects.




# StatefulSets

StatefulSets are the most complicated usage pattern in k8s, thus the most guaranteed to land on your plate, because they are often for legacy applications.
* Employ at-most-one semantics:
	* Ensure reliable scale up/down and deployment properties: one replica is added at a time, sequentially
	* Instances are always added and removed by highest ordinal index
	* Replica creation is more constrained in terms of precise conditions (think unknown states: bad health/liveness probe behavior)
* Have stable hostname, e.g. "stateful-app-0", "stateful-app-1" etc.
* Are often coupled to headless services, such that specific backing pods (leaders, readers, writers, etc, in some pattern) can be reached deterministically
* Use PVC templates for deterministic pv configuration/assignment. Note this implies scheduling constraints on where/how StatefulSets can be run according to their requirements.
* Use sidecar pods to implement and encapsulate; there are examples of this online, for instance using an init container or sidecar to perform leader election among stateful pods.



# Services

### Testing
	`kubectl proxy --port=8888`

Install dig, nslookup or other tools in alpine containers:

	apk update && apk add bind-tools
	# buntu/debian:
	apt update && apt add dnsutils

Naming convention for FQDNs is:
	
	my-app.default.svc.cluster.local
	[app name].[namespace].svc.[cluster]

### Headless
Headless services allow connecting to specific pods backing the service (e.g. statefulset pods), other stable-identity patterns, or even all pods.

Create headless services by setting `ClusterIP: None`. This will create dns A records, then view the records:

	nslookup some-service
	dig SRV some-service

### DNS



### Commands

	# create a service on port 80 for deployment 'mydeployment' for container ports 8080
	kubectl expose deploy/mydeployment --port=80 --container-port=8080

	# serve seure api proxy on port 4242, so you can query it from your local client without auth
	kubectl proxy --port=4242

	# list all available services' addresses injected at container creation
	kubectl exec some-pod env

	nslookup some-headless-service
	dig SRV my-app.default.svc.cluster.local

### Service Gating

Awaiting a service in an init container to latch main container startup:
	
	while true;
		echo "Awaiting $service_name...";
		wget http://$service_name -q -T 1 -o /dev/null >/dev/null 2>/dev/null && break; 
		sleep 1; 
	done;
	echo "$service_name is up";



# RAFT
* etcd uses RAFT, which requires 3, 5, or 7 nodes to maintain quorum.
* etcd instances know of eachother manually via a list of configured peers
* API server sits in front of etcd and statelessly represents its info
* Controllers and Scheduler rely on leader election, since only one of each type may operate to prevent race conditions.

### Commands

	etcdtl put foo bar
	etcdctl del foo
	# Watch for changes to key 'foo'
	etcdctl watch foo 
	etcdctl get foo
	etcdctl get foo --print-value-only
	# Get all keys prefixed 'foo': foo1, foo, foo3, etc.
	etcdctl get --prefix=foo

* Note: each etcd key has a version attached. Use `--rev=n` to get specific versions. They can also use time-bound leases.


# Configuration and Secrets

	kubectl create configmap myconfig --from-literal=foo=bar --from-file=some-cfg.txt  --from-file=./some-directory

This create a config map 'myconfig' using the literal and file syntax. The first file syntax stores the filename as the key and the file content as the value; the latter using a directory stores the sub files as keys and their content as the corresponding values.

### Mounted configmaps

Mounting configmap as a volume allows updating mounted values dynamically for tasks like leader election. Updates depend on the underlying state storage (etcd, sqlite, etc.) Best to look this up as needed (and test like hell to ensure reliability), I just find it useful. More advanced mechanisms likely exist, my info is 2017. See docs:

    spec:
      containers:
      - image: nginx:alpine
        name: web-server
        volumeMounts:
        - name: config
          mountPath: /etc/nginx/conf.d
          readOnly: true
          ...
      volumes:
      - name: config
        configMap:
          name: nginx-config
		
Then update the config and poll to detect the change:

    kubectl edit configmap nginx-config
	kubectl exec nginx -c main cat /etc/nginx/conf.d



### Secrets

Secrets are written to tempfs inside containers, and never persisted to disk.

	openssl genrsa -out https.key 2048
	openssl req -new -x509 -key https.key -out https.cert -days 3650 -subj /CN=www.mysite.com
	kubectl create secret tls site-secret --cert=https.cert --key=https.key
	kubectl get secret site-secret -o yaml

















