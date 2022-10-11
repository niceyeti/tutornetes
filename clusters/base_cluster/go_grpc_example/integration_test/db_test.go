//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	ep "go_grpc_example/endpoints"
	pb "go_grpc_example/proto"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	//"google.golang.org/grpc/status"
)

const (
	DB_USER   = "niceyeti"
	DB_PASSWD = "knockknock"
	addr      = "127.0.0.1:80"
)

var (
	db     *sql.DB
	client pb.CrudServiceClient
)

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_PASSWORD=" + DB_PASSWD,
			"POSTGRES_USER=" + DB_USER,
			"POSTGRES_DB=" + ep.DBName,
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		DB_USER,
		DB_PASSWD,
		hostAndPort,
		ep.DBName)

	log.Println("Connecting to database on url: ", databaseUrl)

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Set up grpc client
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v\n", err)
	}
	defer conn.Close()
	client = pb.NewCrudServiceClient(conn)

	// Set up server

	// Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	if err := conn.Close(); err != nil {
		log.Println(err)
	}

	os.Exit(code)
}

func TestCreate(t *testing.T) {
	log.Println("createPost was invoked")

	post := &pb.Post{
		Id:          "321123",
		AuthorId:    "Jose",
		Title:       "Gone With the Wind",
		Description: "Humpty dumpy",
		FullText:    "In the beginning...",
	}

	res, err := client.CreatePost(context.Background(), post)
	//logErr(err)
	if err != nil {
		t.Fatalf("uhohs! %v", err)
	}

	log.Printf("CreatePost response: %v\n", res)
}
