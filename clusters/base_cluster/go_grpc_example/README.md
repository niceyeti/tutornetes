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