# Notes on a Minimial Distributed State Service

Gist: to gain some traction with gRPC I was considering building a distributed-state application,
since that is a hard problem and can use gRPC in a variety of ways:
* the frontend api: read/write current state
  * FUTURE: a 'watch' functionality using server-push
* the backend quorum api: ensuring state consistency across replicas

The state-app would be dirt simple, although distributed implmentation is a bear:
* there is a single current state S* defined as a raw string
* n replicas all have a copy of S* and communicate with eachother via gRPC to achieve consistency
* in addition to S*, every replica maintains the previous k states (where k is a configurable value)

Use-cases: mainly simple discrete games. Distributing the state provides resilience, and could represent any process for which only the current picture is needed.

Distributed state is not for the custom minded, and luckily is well-settled.
The seemingly 'simple' implementations of things like eventual consistency are not so at all in practice.
Look up eventual-consistency papers and CAP theorems.
A sturdy grasp of these theorems is required, mainly because distributed state is such an 'and also' game
of missing coverage, properties, etc.







