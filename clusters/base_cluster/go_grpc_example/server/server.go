package main

import (
	"context"
	"fmt"
	"log"

	pb "go_grpc_example/proto"

	empty "github.com/golang/protobuf/ptypes/empty"
	"gorm.io/gorm"
)

type Server struct {
	db *gorm.DB
	pb.UnimplementedCrudServiceServer
}

func (s *Server) CreatePost(ctx context.Context, post *pb.Post) (*pb.PostID, error) {
	log.Printf("CreatePost invoked\n")

	dto := NewPost(post)
	tx := s.db.
		WithContext(ctx).
		Create(&dto)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &pb.PostID{
		Id: dto.PostId,
	}, nil
}

func (s *Server) ReadPost(ctx context.Context, postID *pb.PostID) (*pb.Post, error) {
	log.Printf("ReadPost invoked\n")

	post := &Post{}
	tx := s.db.
		WithContext(ctx).
		Where("post_id = ?", postID.Id).
		First(&post)
	if tx.Error != nil {
		log.Printf("error in ReadPost: %v\n", tx.Error)
		return nil, tx.Error
	}

	pbPost := NewPbPost(post)
	return &pbPost, nil
}

// UpdatePost
func (s *Server) UpdatePost(ctx context.Context, pbPost *pb.Post) (*empty.Empty, error) {
	log.Printf("UpdatePost invoked\n")

	post := NewPost(pbPost)
	dest := &Post{}
	tx := s.db.
		WithContext(ctx).
		Where("post_id = ?", post.PostId).
		First(dest)
	if tx.Error != nil {
		log.Printf("error in UpdatePost: %v\n", tx.Error)
		return &empty.Empty{}, tx.Error
	}

	post.ID = dest.ID
	if !Merge(&post, dest) {
		// No changes received, so just return
		log.Println("no post changes in UpdatePost, returning")
		return &empty.Empty{}, nil
	}

	log.Println("new desc: " + fmt.Sprintf("%d ", dest.ID) + dest.Description)
	tx = s.db.
		WithContext(ctx).
		Save(dest)

	s.db.First(dest)

	log.Printf("after update, got: %+v\n", dest)

	return &empty.Empty{}, tx.Error
}

func (s *Server) DeletePost(ctx context.Context, postID *pb.PostID) (*empty.Empty, error) {
	tx := s.db.
		WithContext(ctx).
		Where("post_id = ?", postID.Id).
		Delete(&Post{})

	return &empty.Empty{}, tx.Error
}

// ListPosts streams all of the posts.
// Obviously this could take a where-type clause or other query, omitted for simplicity.
func (s *Server) ListPosts(_ *empty.Empty, lps pb.CrudService_ListPostsServer) error {

	rows, err := s.db.
		Model(&Post{}).
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	// TODO: this function is incomplete, I think I'm missing some stream reqs or closure reqs (or maybe above as well)
	post := &Post{}
	for rows.Next() {
		if err := s.db.ScanRows(rows, post); err != nil {
			return err
		}

		pbPost := NewPbPost(post)
		if err := lps.Send(&pbPost); err != nil {
			return err
		}
	}

	return nil
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
