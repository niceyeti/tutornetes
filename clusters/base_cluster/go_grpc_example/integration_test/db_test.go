// TODO: add build tags here. Currently a tag here disables gopls and intellisense, not sure why.

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	ep "go_grpc_example/endpoints"
	pb "go_grpc_example/proto"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DB_USER   = "niceyeti"
	DB_PASSWD = "knockknock"
	SVC_ADDR  = "127.0.0.1:8080"
)

var (
	db     *sql.DB
	client pb.CrudServiceClient
)

func TestMain(m *testing.M) {
	log.Println("Setting up test resources")

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

	dbAddr := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		DB_USER,
		DB_PASSWD,
		dbAddr,
		ep.DBName)

	log.Println("Connecting to database on url: ", databaseUrl)
	log.Println("Using dbAddr: " + dbAddr)

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

	// Set up the gRPC service
	cfg := ep.AppConfig{
		Addr: SVC_ADDR,
		Cert: "",
		Key:  "",
		DbCreds: ep.DBCreds{
			DbName: ep.DBName,
			Addr:   dbAddr,
			User:   DB_USER,
			Pass:   DB_PASSWD,
		},
	}

	db, err := ep.Connect(&cfg.DbCreds)
	if err != nil {
		log.Fatalf("db connection failed: %v\n", err)
	}

	// Delete the existing db/tables, if any
	ep.DeleteDb(db, cfg.DbCreds.DbName, ep.PostsTable)

	if err = ep.EnsureDB(db, cfg.DbCreds.DbName, &ep.Post{}); err != nil {
		log.Fatalf("%s db creation failed: %v\n", cfg.DbCreds.DbName, err)
	} else {
		log.Printf("%s db exists\n", cfg.DbCreds.DbName)
	}

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	log.Printf("Listening at %s\n", cfg.Addr)

	opts := []grpc.ServerOption{}
	gs := grpc.NewServer(opts...)
	ep := ep.NewServer(db)
	pb.RegisterCrudServiceServer(gs, ep)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		if err := gs.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v\n", err)
		}
	}()
	wg.Wait()

	// Set up grpc client
	conn, err := grpc.Dial(SVC_ADDR, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v\n", err)
	}
	defer conn.Close()
	client = pb.NewCrudServiceClient(conn)

	// Run tests
	code := m.Run()

	gs.GracefulStop()

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
	t.Log("blah")
	post := &pb.Post{
		Id:          "321123",
		AuthorId:    "Jose",
		Title:       "Gone With the Wind",
		Description: "Humpty dumpy",
		FullText:    "In the beginning...",
	}

	res, err := client.CreatePost(context.Background(), post)
	log.Printf("CreatePost response: %v\n", res)
	//logErr(err)
	if err != nil {
		t.Fatalf("uhohs! %v", err)
	}

	log.Printf("CreatePost response: %v\n", res)
}
