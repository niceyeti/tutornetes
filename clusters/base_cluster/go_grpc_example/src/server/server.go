package main

import pb "go_grpc_example/src/proto"

type Server struct {
	pb.CrudServiceServer
}
