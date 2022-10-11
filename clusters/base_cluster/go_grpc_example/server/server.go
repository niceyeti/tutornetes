package main

import (
	"context"
	"log"
	"strconv"

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
	dto.ID = 353
	// TODO: review gorm docs and convention, I'm flying by the seat of my pants. ID should (?) autoincrement.
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
		Where("id = ?", postID.Id).
		First(&post)
	if tx.Error != nil {
		log.Printf("error in ReadPost: %v\n", tx.Error)
		return nil, tx.Error
	}

	pbPost := NewPbPost(post)
	return &pbPost, nil
}

// UpdatePost
func (s *Server) UpdatePost(ctx context.Context, post *Post) (*empty.Empty, error) {
	dest := &Post{}
	tx := s.db.First(dest, post.ID)
	if tx.Error != nil {
		return &empty.Empty{}, tx.Error
	}

	// No changes needed, just return
	if !Merge(post, dest) {
		return &empty.Empty{}, nil
	}

	tx = s.db.
		WithContext(ctx).
		Save(dest)
	return &empty.Empty{}, tx.Error
}

func (s *Server) DeletePost(ctx context.Context, postID *pb.PostID) (*empty.Empty, error) {
	id, err := strconv.ParseUint(postID.Id, 10, 32)
	if err != nil {
		return &empty.Empty{}, err
	}

	tx := s.db.
		WithContext(ctx).
		Delete(&Post{}, id)
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
