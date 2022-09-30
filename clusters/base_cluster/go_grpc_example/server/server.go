package main

import (
	pb "go_grpc_example/proto"
)

type Server struct {
	//pb.CrudServiceServer
	pb.UnimplementedCrudServiceServer
}

/*
func (*Server) (ctx context.Context, req *pb.SqrtRequest) (*pb.SqrtResponse, error) {
	log.Printf("Sqrt was invoked with number %d\n", req.Number)

	number := req.Number

	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %d", number),
		)
	}

	return &pb.SqrtResponse{
		Result: math.Sqrt(float64(number)),
	}, nil
}
*/
