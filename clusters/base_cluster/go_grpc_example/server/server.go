package main

import (
	"context"
	pb "go_grpc_example/proto"
	"log"
	"strconv"

	"gorm.io/gorm"
)

type Server struct {
	db *gorm.DB
	pb.UnimplementedCrudServiceServer
}

func (s *Server) CreatePost(ctx context.Context, post *pb.Post) (*pb.PostID, error) {
	log.Printf("CreatePost invoked\n")

	dto := NewPost(post)
	// TODO: review gorm docs and convention, I'm flying by the seat of my pants. ID should (?) autoincrement.
	dto.ID = 13
	tx := s.db.
		WithContext(ctx).
		Create(&dto)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &pb.PostID{
		Id: strconv.FormatUint(uint64(dto.ID), 10),
	}, nil
}

func (s *Server) ReadPost(ctx context.Context, postID *pb.PostID) (*pb.Post, error) {
	log.Printf("ReadPost invoked\n")

	post := &Post{}
	tx := s.db.
		WithContext(ctx).
		Where("post_id = ?", postID).
		First(&post)
	if tx.Error != nil {
		log.Printf("error in ReadPost: %v\n", tx.Error)
		return nil, tx.Error
	}
	pbPost := NewPbPost(post)

	return &pbPost, nil
}

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
