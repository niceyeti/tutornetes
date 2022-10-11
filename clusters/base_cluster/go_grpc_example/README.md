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
    * export DB_USER=niceyeti; export DB_PASSWORD=niceyeti; export DB_HOST=172.17.0.1; export DB_PORT=5432; export DEV=true; ./bin/service
4) Debug using this:
    * DB_USER=niceyeti DB_PASSWORD=niceyeti DB_HOST=172.17.0.1 DB_PORT=5432 dlv debug main.go

### Smoother Workflow

Due to the fact that the database contains state, although one could deploy a postgres container
in the cluster and develop by setting up a tiltfile, its actually somewhat easier for the sake of a
mere demo to develop locally using the manual steps above for starters, then using ORY to develop
using integration-driven development. The complexity is due to:
1) db state
2) client and server testing

In most cases an external db carries with it its own custom requirements and coordination with its dba.

Test strategy:
1) factor out as many interfaces as possible to pull them into fast unit tests
2) integration testing with the db: annotate the integration test file so it can be toggled via build flags.

### The Database

The database is merely for demonstration purposes and playing with gorm,
and describes CRUD operations for a bunch of 'Post' objects which are like blog posts.

NOTE: the code here is not batteries-included.
It is not free from sql-injection, has no fluent validation checks, nor did I fully review the gorm docs.
There could be much to gain in terms of cleaner implementation, layering, security, connection management, and so on.

#### Time
Time is highly important in a real database, whereas I am simply using time.Time fields of gorm.
Still, you always want to know the impact of the types of time fields used, 8601/3339 format considerations,
monotonic times, and the timezone of the database (which also may or may not be implemented in this code).
If adapting this code, review time considerations for maintainability and correctness.

#### GORM
I like gorm, it makes interacting with databases 'easy'. Is my attitude sufficient for production? Nope.
Do more research on reviews of gorm and whether or not it works for more complex databases before using.

Also clearly implement the distinction between soft and hard deletion.
And clearly define the data flow of the exposed objects: should the grpc api expose the object
ids generated by the db, or some other id?

## Testing

Run:
* `go test ./integration_test/ -tags integration`
Note this has to be run from the host since dockertest uses docker and I'm not going to add docker to the dev container.





### Diagnostics and Tools

Port pings:
* nmap 127.0.0.1 -p 5432
* telnet 127.0.0.1 5432

### Extensions and Futures

These are purely ideas for practice/job-prep.
- Write and add a cache to the service
- Refactor apps layers and compare with other grpc app layouts
- kube-ify the app, with kubes based tests. Basically try to develop the most advanced and smooth
  devops workflow using tilt and by fully parameterizing the application wrt the db and so on.
- Implement all timing requirements (grpc timeouts, grpc.ServerOptions, etc). None are specified.
