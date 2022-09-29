# gRPC Notes

Notes on a gRPC mini-course.

### Fluff

* Stack: gRPC -> protocol buffers -> HTTP2
* Design: gRPC allows for more expressive api's than http-verb REST apis
* Security: gRPC is secured by default using SSL, whose params must be passed in at connection creation.
* References: https://grpc.io/docs/what-is-grpc/core-concepts/

gRPC is beneficial at the most basic level because it is a binary protocol, and thus uses less data and is less cpu-intensive since the mapping to/from binary is simpler than json for example.
* more compressed, less data
* less cpu burden
There are significant performance benefits using gRPC vs traditonal HTTP/REST.

In addition, it runs over HTTP2, and thus inherits its benefits:
* continuous tcp connections
* better security compared with multiple transactions based on multiple tcp connections

| | HTTP | HTTP2 |
|----|-----|------|
| connections | one per request | reuses connection (multiplexed) |
| headers | not compressed | compressed |
| server push| not possible | possible | 
| security | more chatter, less secure | less chatter, more secure, ssl required |

Scalability: server is async, client is either async or blocking.

### RPC Api Types

Unary: 1:1 client/server transactions. This is the most similar to traditional REST request/response apis.
Server streaming: client sends one request and server sends multiple responses.
    * *Note: the client reads responses until it receives EOF, which ends the communication.*
Client streaming: client sends one or multiple requests, server sends a single response.
    * *Note: the server receives messages from the client until it receives EOF, and **then** sends its response and closes.
Bidirectional: both client and server can send and receive multiple requests/responses over the same multiplexed connection. This is basically two go routines, one calling stream.Send() and the other calling stream.Recv() and coordinating shutdown on EOF or errors.

NOTE: streams are not just for persistent streams, but also for collection calls, i.e. `ListBlogs(google.proto.empty) returns (stream Blog)`.

Note how the client/server models differ. In server streaming, the client reads until there is no more input, whereas in client streaming the server reads until EOF and then sends an aggregated response. Therefore, gRPC fits the following api use-cases:
* unary transactions: traditional REST-like apis
* stream messages continuously: continuous information, realtime apis
* stream messages and aggregate a single response: uploading, aggregate computations, etc
* bidirectional streaming: as described above, just two goroutines where one calls Send() and the other calls Recv(). 

```
    service GreetService {
        // unary
        rpc Greet(GreatRequest) returns (GreetResponse) {}
        
        // server streaming
        rpc GreetStream(GreetRequest) returns (stream GreetResponse) {}

        // client streaming
        rpc LongGreat(stream GreetRequest) returns (GreetResponse) {}

        // bidirectional streaming
        rpc GreetManyToMany(stream GreetRequest) returns (stream GreetResponse) {}
    }
```

### Maintenance and Testing



### gRPC Dependencies and Requirements

See the gRPC site, but the gist is you need the protobuf-compiler cmd line tool and the golang libs. In a Dockerfile, these are:

```
# Install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends protobuf-compiler
...
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

Protobuf usage for generating code and compiling a client/server pair
* Define your api in a .proto file
* ```protoc -I$@/${PROTO_DIR} --go_opt=module=${PACKAGE} --go_out=. --go-grpc_opt=module=${PACKAGE} --go-grpc_out=. $@/${PROTO_DIR}/*.proto```
* ```go build -o ${BIN_DIR}/$@/${SERVER_BIN} ./$@/${SERVER_DIR}```
* ```go build -o ${BIN_DIR}/$@/${CLIENT_BIN} ./$@/${CLIENT_DIR}```

### Basic Workflow

1) Define .proto files, such as in a `proto/` folder. In the .proto file, specify the path to the proto folder:
    * ```option go_package = "github.com/bob/proto-example/proto"```
2) Compile them with protoc:
* protoc -I myproject/proto --go_opt=module=github.com/bob/proto-example --go_out=. --go-grpc_opt=module=github.com/bob/proto-example --go-grpc_out=. myproject/proto/*.proto
3) Find the api functions in your generated code, copy their definitions and refer to these in your go code, and build.

### Api Usage

There are two usages of gRPC, unary and streaming, directional w.r.t. whether the client or server is streaming. Unary gRPC is straightforward. Streaming from server to client will successively send a message definition (some struct). The client breaks when it receives EOF from `stream.Recv()`.

Server:
```
func (s *Server) GreetManyTimes(in *pb.GreetRequest, stream pb.GreetService_GreetManyTimesServer) error {
    for i := 0; i < 10; i++ {
        stream.Send(&pb.GreetResponse{
            Result: fmt.Sprintf("Hello %d times", i),
        })
    }
}
```
Client:
```
func doGreetManyTimes(ctx context.Context, c pb.GreetServiceClient) {
    req := &pb.GreetRequest{
        Name: "John",
    }
    stream, _ := c.GreetManyTimes(ctx, req)
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            break
        }
    }
}
```

### Errors and Deadlines

The gRPC golang libraries include facilities for returning gRPC-specific errors, or for determining
if an error is a gRPC error or some other error, using grpc/codes and grpc/status:
* status.Errorf: used to return gRPC errors (invalid api usage, etc.)
* status.FromError: used to determine if an error is a gRPC error, for error mapping/handling

Each has their use-cases, depending on one's api requirements. Of course, well-defined errors should always by first-class members in an api's definition, so these functions should be used in most code.

Deadlines: the generated endpoint signatures all accept a Context parameter; this can be used as needed 
to check for deadline-exceeded, etc. Clients implement the context deadline by simply passing their own context (with-timeout, for example) to their api call. Clients can likewise detect if a returned err
is a result of context expiration using `err.Code() == codes.DeadlineExceeded`. Thus context/deadline handling also uses library error handling functions.

### Reflection

One can modify ones gRPC server code to implement reflection using golang.google.com/grpc/reflection package, and a third party library called 'evans'. This library allows interacting with one's server
using the command line to view message definitions, api, and calling functions. This can be used for automation, testing, and debugging.

### Generic Api Development Outline

1) Define .proto file with data types and service api.
2) Compile this, then import the code and implement the service method definitions.

### Resources

A good overview of developer concerns, including gRPC healthchecks:
* https://milad.dev/posts/grpc-in-microservices/