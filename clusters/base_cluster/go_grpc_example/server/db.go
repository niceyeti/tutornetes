package main

import (
	"fmt"
	"log"
	"time"

	pb "go_grpc_example/proto"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: move this file to db layer in separate package

// DBCreds describes the connection creds to the database, which should not be persisted
// and must be transmitted securely (e.g. via Kubernetes secrets).
type DBCreds struct {
	DbName string
	Addr   string
	User   string
	Pass   string
}

// Post is just for demonstration, and such a simple object hardly describes the
// relational and other capabilities of gorm and sql. For example, every field
// here might be decorated with indices, constraints, or read/write CRUD restrictions.
// The decorations specified below 'just work' and are oterhwise incomplete.
type Post struct {
	// Note: explicitly implementing these fields seems better than embedding gorm.Model. For one,
	// this ensure the ID autoincrements, rather than requiring the service to generate ids.
	ID        uint `gorm:"primaryKey;autoIncrement:true;unique" sql:"AUTO_INCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	// PostId is redundant wrt to ID, but seems about right to hide the internal id by default.
	// Id's represent something to their consumer, in this case the app layer; hence it could be
	// something like the hash of post fields, a concatenation of logical ones, whatever ones reqs.
	PostId      string `json:"post_id,omitempty"`
	AuthorId    string `json:"author_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	FullText    string `json:"full_text,omitempty"`
}

const (
	DBName          = "blog"
	PostsTable      = "posts"
	DB_HOST         = "DB_HOST"
	DB_PORT         = "DB_PORT"
	DB_USER         = "DB_USER"
	DB_PASSWORD     = "DB_PASSWORD"
	DB_HOST_DEFAULT = "127.0.0.1"
	DB_PORT_DEFAULT = "5432"
	DB_USER_PATH    = "/etc/secrets/db/user"
	DB_PASS_PATH    = "/etc/secrets/db/passwd"
)

func NewPost(pbPost *pb.Post) Post {
	return Post{
		PostId:      pbPost.Id,
		AuthorId:    pbPost.AuthorId,
		Title:       pbPost.Title,
		Description: pbPost.Description,
		FullText:    pbPost.FullText,
	}
}

func NewPbPost(post *Post) pb.Post {
	return pb.Post{
		Id:          post.PostId,
		AuthorId:    post.AuthorId,
		Title:       post.Title,
		Description: post.Description,
		FullText:    post.FullText,
	}
}

func ReadDBConfig() (*DBCreds, error) {
	dbHost := getEnv(DB_HOST, DB_HOST_DEFAULT)
	dbPort := getEnv(DB_PORT, DB_PORT_DEFAULT)

	var err error
	dbUser := getEnv(DB_USER, "")
	if dbUser == "" {
		dbUser, err = getTrimmedConfig(DB_USER_PATH, "")
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Warning: db cred taken from insecure env. In prod, creds should be transferred via tempfs instead.")
	}

	dbPass := getEnv(DB_PASSWORD, "")
	if dbPass == "" {
		dbPass, err = getTrimmedConfig(DB_PASS_PATH, "")
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Warning: db cred taken from insecure env. In prod, creds should be transferred via tempfs instead.")
	}

	return &DBCreds{
		Addr:   fmt.Sprintf("%s:%s", dbHost, dbPort),
		User:   dbUser,
		Pass:   dbPass,
		DbName: DBName,
	}, nil
}

// Connect returns a gorm.DB for the passed creds.
func Connect(creds *DBCreds) (*gorm.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s", ///posts?sslmode=disable",
		creds.User,
		creds.Pass,
		creds.Addr)
	return gorm.Open(
		postgres.New(
			postgres.Config{
				DSN:                  dsn,
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			}), &gorm.Config{})
}

func DeleteDb(db *gorm.DB, dbName string) {
	log.Println("WARNING: deleting existing db, if it exists. This is only for development.")
	tx := db.Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
	if tx.Error != nil {
		log.Printf("DeleteDB error: %v\n", tx.Error)
	}
}

// EnsureDB checks if a database exists by attempting to query the passed table.
// This function is purely for development and is not a robust way to check.
func EnsureDB(db *gorm.DB, dbName string, migrateObj interface{}) error {
	// TODO: what is this 'sql injection' of which thou speak?
	tx := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
	if tx.Error != nil {
		log.Printf("%v\n", tx.Error)
	}

	// Migrate the schema
	return db.AutoMigrate(migrateObj)
}
