# gRPC CRUD App

A sample grpc CRUD api backed by SQL in golang. 
In most cases such a microservice would interact only with other gRPC
backend services, in front of which is some http-based gateway-api.
This app merely satisfied a few training goals in one:
1) Build a basic gRPC api
2) Back an app with postgres in a k8s cluster
3) Implement a bare-bones integration test/workflow with dockertest

The (unary) transactional nature of a CRUD api means this only uses a subset of the capability of gRPC.
Oh well.

### gRPC Use-Cases

gRPC is good at replacing existing REST apis, which can be done using the simplest unary-gRPC api definition. It is also amenable to more dynamic cases of message passing using streaming.
The streaming model runs over tcp/HTTP2 and facilitates many data sharing use-cases, especially
for confidentiality, but for something like gaming, VOIP, or video (e.g. using UDP) it would only be useful for discrete application models where state can be represented as compact transactions.


### Lessons Learned

A generic gRPC api to some other backend db or service is really just a bunch of boilerplate,
and could almost yield itself to code-generation if specified within some constraints. This
provides a good way to think about such api's as mere layers in an infrastructure; and
they should be trivialized as such. Focus instead on other aspects of technical debt.

After completing this app, I almost err on the side of not adding gRPC endpoints to a platform unless they are native to it and well-understood by other developers and maintainers. In other words gRPC requires a prior organizational/team commitment.
 This may be an overstatement, and should simply be evaluated on a case-by-case basis. However, integrating gRPC/protoc into development and maintenance is a prime suspect for maintenance headaches for team mates without gRPC experience or who are innocently unfamiliar with how it was set up in one's repo. The test, 
development, and maintenance plan must be defined, well-understood, and agreed upon by stakeholders.
The point is that gRPC, though great, anticipates justified 'why have you added this?' feedback from feature-slaying, resource-conscientious managers :) . Though let's not forget gRPC's primary features
w.r.t. http2, security, and performance. Some feedback:
* don't forget locking in your service (at this writing, the gRPC service has none)
* simplify, simplify, simplify, eliminate, eliminate, eliminate:
    * I would estimate that merely completing the initial gRPC service skeleton and barebones integration
      test took about 4x longer than I mentally estimated. This was due to the usual trip-ups: poor docs in test/code libs, added test complexity, incremental (aka timid) development, and so on.
    * while gRPC endpoints could almost be generated somehow, every line still needs to be tested.
    * the above simply indicate how much overhead is created by even routine microservice endpoints.
    * and i haven't even added tls/ssl!
* start with integration-driven development using dockertest, if possible. Doing so also means you can
  add pprof metric collection to tests.
* YAGNI! It bears repeating: the foremost responsibility is simplifying or eliminating parts of the api. The design phase should ruthlessly eliminate as much downstream code as possible.
* keep an eye out for more efficient test/development methods and libraries. The integration test is a heavyweight test; golang intuition tells me there are gRPC test libraries and strategies akin to the quality and ease of the httptest library.

### TODO
Grep for TODOs left in the code, these are merely high-level points.
- inject the db into the server to enable separate unit/integration testing. Per go practice, define the db interface the service wishes to consume (a subset of gorm.DB perhaps), and generate mocks.
- for fun, write a full benchmark test with pprof output: use the gRPC client to implement burden testing,
  add pprof code to the benchmark test, and use it to monitor performance.
- locking
- kubeify, dockerfile, tilt, copy from build env to scratch in Dockerfile
- document as if this were a production app: identify stakeholders and responsibilities,
  ie, maintenance and testing.

### Development and Testing

These how-tos are somewhat out of sync, but basically one can develop a gRPC endpoint using this
spiral:
1) implement a makefile that builds both the client and server, and test manually (e.g. before backing with a real db)
2) back the service with an actual postgres container
3) build a full integration test using dockertest

Postgres installation for testing (change params appropriately):
* `docker pull postgres:latest`
* `docker run -itd -e POSTGRES_USER=niceyeti -e POSTGRES_PASSWORD=niceyeti -p 5432:5432 -v /data:/var/lib/postgresql/data --name postgresql postgres`
    * NOTE: consider network attachment options to ease reaching the db through various container environments.
      There are natural network connection complications to address whenever running.
* Source: https://www.baeldung.com/ops/postgresql-docker-setup

### Profiling

For a production gRPC app, especially backed by postgres or another external resource,
I would absolutely require having a good profiling workflow in place for my own
development. Memory leaks and performance issues can occur down one's entire stack of
third-party and other components, nor can you assume libraries will perform in complimentary
fashion with one another. Basically, its 3am--do you know where your goroutines are?

I had more complicated ideas for how to profile the gRPC/postgres endpoints at runtime,
however pprof generally has any tools you could need, batteries included, and it integrates with testing.
Other options are available depening on your use-case, but I found that the two most valuable
methods are to simply call `go test` with prof flags, or to compile pprof commands into the integration
test itself. The former gather metrics encompassing the entire test environment, including lots of setup
noise; the latter is better because it allows one to scope what is benchmarked.
1) go test -v ./integration_test/ -memprofile=mem.prof -cpuprofile=cpu.prof -blockprofile=block.prof
2) go tool pprof -http=":8080" mem.prof

With infinite leisure time, one could also compile and run the pprof http server into one's gRPC server code, run the server and query it in various ways, while monitoring stats from the pprof server.
The gRPC-client could be used to automate various queries for ad-hoc benchmarking and system burden testing.
This method could be implemented entirely within a benchmark integration-test, but doing so probably
would not be supported by organizational resources.

Concluding, just be familiar with the pprof tooling and how to interpret its outputs.
Building one's code to facilitate analysis is also helpful, such as a client-driven development
approach whereby one incrementally builds a client capable of more complex load testing.

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
3) dockertest requires running docker in the cluster or the vscode container, which is complexity I don't want

In most cases an external db carries with it its own custom requirements and coordination with its dba.

Current development/test workflow:
1) Modify client or server. Small changes can be tested using the manual-workflow steps and the makefile.
2) From /clusters/base_cluster/go_grpc_example, run the integration test:
    * `go test -v ./integration_test/`
    * Requirement: you'll want to run with a good network connection, since the integration test may need pull the postgres image

### The Database

The database is merely for demonstration purposes and playing with gorm,
and describes CRUD operations for a bunch of 'Post' objects which are like blog posts.

NOTE: the code here is not batteries-included.
It is not free from sql-injection, has no fluent validation checks, nor did I fully review the gorm docs.
There could be much to gain in terms of cleaner implementation, layering, security, connection management, and so on.

#### Concurrency and Races
Note that very little consideration was given to concurrency requirements in the service,
since I only test the CRUD interfaces serially, one by one. To use a Kamalism, there are many considerations
that should be considered.

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
- Refactor app layers and compare with other grpc app layouts
- kube-ify the app, with kubes based tests. Basically try to develop the most advanced and smooth
  devops workflow using tilt and by fully parameterizing the application wrt the db and so on.
- Implement all timing requirements (grpc timeouts, grpc.ServerOptions, etc). None are specified.
