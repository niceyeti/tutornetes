package main

import (
	"fmt"
	"log"

	pb "go_grpc_example/proto"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: move to db layer in separate package

type Post struct {
	gorm.Model         // Adds an ID field
	Id          string `json:"id,omitempty"`
	AuthorId    string `json:"author_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	FullText    string `json:"full_text,omitempty"`
}

func NewPost(pbPost *pb.Post) Post {
	return Post{
		Id:          pbPost.Id,
		AuthorId:    pbPost.AuthorId,
		Title:       pbPost.Title,
		Description: pbPost.Description,
		FullText:    pbPost.FullText,
	}
}

// Connect returns a gorm.DB for the passed creds.
func Connect(creds *DBCreds) (*gorm.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s", ///posts?sslmode=disable",
		creds.User,
		creds.Pass,
		creds.Addr)
	return gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
}

func EnsureDB(db *gorm.DB, dbName string) error {
	// check if db already exists
	stmt := fmt.Sprintf("SELECT * FROM %s", dbName)
	tx := db.Raw(stmt)
	// TODO: left error checking here non-robust, but could fix.
	if tx.Error == gorm.ErrInvalidDB || tx.Error == nil {
		return nil
	}

	tx = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
	if tx.Error != nil {
		log.Printf("%t\n", tx.Error)
	}

	return tx.Error
}
