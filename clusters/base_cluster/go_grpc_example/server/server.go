package main

import (
	"context"
	pb "go_grpc_example/proto"
	"log"
)

type Server struct {
	//pb.CrudServiceServer
	pb.UnimplementedCrudServiceServer
}

func (s *Server) CreatePost(ctx context.Context, post *pb.Post) (*pb.PostID, error) {
	log.Printf("CreatePost invoked\n")

	return &pb.PostID{Id: "123"}, nil
}

//func (s *Server) ReadPost(context.Context, *PostID) (*empty.Empty, error)
//func (s *Server) UpdatePost(context.Context, *Post) (*empty.Empty, error)
//func (s *Server) DeletePost(context.Context, *PostID) (*empty.Empty, error)
//func (s *Server) ListPosts(*empty.Empty, CrudService_ListPostsServer) error

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
