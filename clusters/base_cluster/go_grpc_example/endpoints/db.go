package endpoints

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
	ID        uint           `gorm:"primaryKey;autoIncrement;uniqueIndex" json:"id,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	//gorm.Model
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

// Merge updates fields in dest with the non-empty fields of src.
// The return value indicates if any update occurred.
// Afterward, dest will contain the id, post-id, and other mandatory fields of src.
func Merge(src, dest *Post) (updated bool) {
	dest.ID = src.ID
	dest.PostId = src.PostId
	dest.CreatedAt = src.CreatedAt
	dest.DeletedAt = src.DeletedAt
	dest.UpdatedAt = src.UpdatedAt

	if src.AuthorId != "" && src.AuthorId != dest.AuthorId {
		log.Println("updating authorID")
		dest.AuthorId = src.AuthorId
		updated = true
	}
	if src.Description != "" && src.Description != dest.Description {
		log.Println("updating description")
		dest.Description = src.Description
		updated = true
	}
	if src.FullText != "" && src.FullText != dest.FullText {
		log.Println("updating fulltext")
		dest.FullText = src.FullText
		updated = true
	}
	if src.Title != "" && src.Title != dest.Title {
		log.Println("updating title")
		dest.Title = src.Title
		updated = true
	}
	return
}

func ReadDBConfig() (*DBCreds, error) {
	dbHost := GetEnv(DB_HOST, DB_HOST_DEFAULT)
	dbPort := GetEnv(DB_PORT, DB_PORT_DEFAULT)

	var err error
	dbUser := GetEnv(DB_USER, "")
	if dbUser == "" {
		dbUser, err = GetTrimmedConfig(DB_USER_PATH, "")
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Warning: db cred taken from insecure env. In prod, creds should be transferred via tempfs instead.")
	}

	dbPass := GetEnv(DB_PASSWORD, "")
	if dbPass == "" {
		dbPass, err = GetTrimmedConfig(DB_PASS_PATH, "")
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
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s", ///posts?sslmode=disable",
		creds.User,
		creds.Pass,
		creds.Addr,
		creds.DbName)
	log.Println("Connecting to dsn " + dsn)

	return gorm.Open(
		postgres.New(
			postgres.Config{
				DSN:                  dsn,
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			}), &gorm.Config{})
}

func DeleteDb(db *gorm.DB, dbName, tableName string) {
	log.Println("WARNING: deleting existing table, if it exists. This is only for development.")
	tx := db.Exec(fmt.Sprintf("DROP TABLE %s;", tableName))
	if tx.Error != nil {
		log.Printf("DeleteDB dropping table %s: %v\n", tableName, tx.Error)
	}

	log.Println("WARNING: deleting existing db, if it exists. This is only for development.")
	tx = db.Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
	if tx.Error != nil {
		log.Printf("DeleteDB error dropping db %s: %v\n", dbName, tx.Error)
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
