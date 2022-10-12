// TODO: add build tags here. Currently a tag here disables gopls and intellisense, not sure why.
// Notes: this is merely the framework for test-driven development of the CRUD interface. These
// tests are likely very fragile, since convey may run tests in parallel, and likewise these tests
// rely on db state. 80/20: the goal here is to exercise the core CRUD functions, but need not
// be complete, nor robust, especially since the api itself is not fully factored/specified
// in terms of things like error-handling, timing, and so on. These tests themselves may
// require locking around the client to ensure no conflicts between concurrent tests.
// Test coverage is far from complete:
// - happy paths complete?
// - exercise all expected api errors
// - request an unexpected thing
// - request with a canceled context
// - idempotence: request a thing/action twice (create twice, update twice, etc)
// - ListPosts not exercised
// - add a built tag to this test '+integration'. Currently this causes gopls to stop operating on
//   this file, because I didn't spend any time researching the issue (e.g. gopls settings).

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
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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

	fmt.Println("wah?")

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
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

	resource.Expire(20) // Tell docker to hard kill the container in 120 seconds

	// TODO: reset to 120s, and expiration above
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 10 * time.Second
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

	// Start the server and wait for it to be up...
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		if err := gs.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v\n", err)
		}
	}()
	wg.Wait()
	time.Sleep(time.Millisecond * 50)

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

func TestCreatePost(t *testing.T) {
	t.Log("createPost was invoked")
	post := &pb.Post{
		Id:          "321123",
		AuthorId:    "Jose",
		Title:       "Gone With the Wind",
		Description: "Humpty dumpy",
		FullText:    "In the beginning...",
	}

	Convey("CreatePost tests", t, func() {
		res, err := client.CreatePost(context.Background(), post)
		So(err, ShouldBeNil)
		So(res.Id, ShouldEqual, post.Id)
	})

	// FUTURE: ensure duplicate posts cannot be created
	//Convey("CreatePost tests", t, func() {
	//	res, err := client.CreatePost(context.Background(), post)
	//	log.Printf("CreatePost response: %v\n", res)
	//
	//	So(err, ShouldBeNil)
	//	So(res.Id, ShouldEqual, post.Id)
	//})
}

func TestReadPost(t *testing.T) {
	post := &pb.Post{
		Id:          "321123",
		AuthorId:    "Jose",
		Title:       "Gone With the Wind",
		Description: "Humpty dumpy",
		FullText:    "In the beginning...",
	}

	Convey("ReadPost tests", t, func() {
		Convey("When a non-existent post-id is requested", func() {
			resultPost, err := client.ReadPost(context.Background(), &pb.PostID{Id: "junk"})
			So(resultPost, ShouldBeNil)
			// TODO: comparing error strings is poor practice and indicates that the api should be refactored.
			// Leaving as-is because this app is purely a demo. But in practice the api should specify its
			// errors as first-class definitions, e.g. ErrPostNotFound, per golang error best-practices.
			So(err.Error(), ShouldEqual, "rpc error: code = Unknown desc = record not found")
		})

		Convey("When an existing post is requested", func() {
			res, err := client.CreatePost(context.Background(), post)
			log.Printf("CreatePost response: %v\n", res)
			So(err, ShouldBeNil)
			So(res.Id, ShouldEqual, post.Id)

			postId := &pb.PostID{
				Id: res.Id,
			}

			Convey("When a post is requested normally (happy path)", func() {
				resultPost, err := client.ReadPost(context.Background(), postId)
				So(err, ShouldBeNil)
				So(resultPost.Id, ShouldEqual, post.Id)
				So(resultPost.FullText, ShouldEqual, post.FullText)
			})

			Convey("When a post is requested with a cancelled context", func() {
				cancelledCtx, cancelFunc := context.WithCancel(context.Background())
				cancelFunc()
				// Allow cancellation to propagate to the context; this may not be
				// necessary, I would have to review the context code.
				time.Sleep(10 * time.Millisecond)

				res, err := client.ReadPost(cancelledCtx, postId)
				So(res, ShouldBeNil)
				So(err, ShouldNotBeNil)

				code := status.Code(err)
				So(code, ShouldEqual, codes.Canceled)
			})
		})
	})
}

func TestUpdatePost(t *testing.T) {
	post := &pb.Post{
		Id:          "321123",
		AuthorId:    "Jose",
		Title:       "Gone With the Wind",
		Description: "Humpty dumpy",
		FullText:    "In the beginning...",
	}

	Convey("UpdatePost tests", t, func() {
		res, err := client.CreatePost(context.Background(), post)
		So(err, ShouldBeNil)
		So(res.Id, ShouldEqual, post.Id)

		Convey("When the post's description is updated", func() {
			newDesc := post.Description + " " + time.Now().Format(time.RFC3339)
			update := &pb.Post{
				Id:          post.Id,
				AuthorId:    post.AuthorId,
				FullText:    post.FullText,
				Description: newDesc,
				Title:       post.Title,
			}
			_, err := client.UpdatePost(context.Background(), update)
			So(err, ShouldBeNil)

			res, err := client.ReadPost(context.Background(), &pb.PostID{Id: post.Id})
			So(err, ShouldBeNil)
			So(res.Id, ShouldEqual, post.Id)
			So(res.Description, ShouldEqual, newDesc)
		})
	})
}

func TestDeletePost(t *testing.T) {
	post := &pb.Post{
		Id:          "321123",
		AuthorId:    "Jose",
		Title:       "Gone With the Wind",
		Description: "Humpty dumpy",
		FullText:    "In the beginning...",
	}

	Convey("DeletePost tests", t, func() {
		res, err := client.CreatePost(context.Background(), post)
		So(err, ShouldBeNil)
		So(res.Id, ShouldEqual, post.Id)

		_, err = client.DeletePost(context.Background(), &pb.PostID{Id: post.Id})
		So(err, ShouldBeNil)
	})
}
