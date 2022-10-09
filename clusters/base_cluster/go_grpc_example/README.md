# gRPC CRUD App

A sample grpc CRUD api backed by SQL in golang. 
In most cases such a microservice would interact only with other gRPC
backend services, in front of which is some http-based gateway-api.
This app merely satisfied a few training goals in one:
1) Build a basic gRPC api
2) Back an app with postgres in a k8s cluster

The (unary) transactional nature of CRUD means this only uses about 25% of the capability of gRPC.
Oh well.

### gRPC Use-Cases

gRPC is good at replacing existing REST apis, which can be done using the simplest unary-gRPC api definition. It is also amenable to more dynamic cases of message passing using streaming.
The streaming model runs over tcp/HTTP2 and facilitates many data sharing use-cases, especially
for confidentiality, but for something like gaming, VOIP, or video (e.g. using UDP) it would only be useful for discrete application models where state can be represented as compact transactions.


### Lessons Learned

A generic gRPC api to some other backend db or service is really just a bunch of boilerplate,
and could almost yield itself to code-generation if specified with some constraints. This
provides a good way to think about such api's as mere connectors in an infrastructure; and
they should be trivialized as such. Focus instead on other aspects of technical debt.

### TODO
- copy from build env to scratch in Dockerfile
- automated testing for project
- db migrations, etc
- document as if this were a production app: identify stakeholders and responsibilities,
  ie, maintenance and testing.

### Development and Testing

Postgres installation for testing (change params appropriately):
* `docker pull postgres:latest`
* `docker run -itd -e POSTGRES_USER=niceyeti -e POSTGRES_PASSWORD=niceyeti -p 5432:5432 -v /data:/var/lib/postgresql/data --name postgresql postgres`
    * NOTE: consider network attachment options to ease reaching the db through various container environments.
      There are natural network connection complications to address whenever running.
* Source: https://www.baeldung.com/ops/postgresql-docker-setup

### Manual Development Workflow

Note: this is subject to change, the makefile workflow is kludgy and not amenable to k8s dev yet.
However the makefile is a nice way to elucidate how the app should be built, and can then
be translated into a tiltfile, helm yaml, dockerfile, and so on for k8s deployment/development.

1) Make any code changes and run `protoc` to generate the code interfaces to be implemented by the client and server.
    * See makefile: `make cproto`
2) Copy the generated interfaces to your server or client code and implement them.
3) Build the client and server:
    * `make all` or `make crud`

Developing with the db, on host and no k3d cluster:
1) Run the db (from host, not vs code container):
    * `docker run -itd -e POSTGRES_USER=niceyeti -e POSTGRES_PASSWORD=niceyeti -p 5432:5432 -v /data:/var/lib/postgresql/data --name postgresql postgres`
2) Run `make all`
3) Run the server (lazy way with env vars):
    * export DB_USER=niceyeti; export DB_PASSWORD=niceyeti; export DB_HOST=172.17.0.1; export DB_PORT=5432; ./bin/server

### Diagnostics and Tools

Port pings:
* nmap 127.0.0.1 -p 5432
* telnet 127.0.0.1 5432