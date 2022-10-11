package main

import (
	"log"
	"net"
	"os"

	ep "go_grpc_example/endpoints"
	pb "go_grpc_example/proto"

	// Note: there is little to no reason to use viper in this app, I wanted to play with it.
	// It does have terrific functions for reading and monitoring configs locally and remotely
	// for interesting use-cases like hot reloads; none of that is needed in this app.

	"google.golang.org/grpc"
)

// FUTURE: a bit awkward, the meat of main could live in the endpoints folder, however
// this approximates an ideal main that binds together pkgs for the db, controller, and config.
func main() {
	cfg, err := ep.ReadAppConfig()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	db, err := ep.Connect(&cfg.DbCreds)
	if err != nil {
		log.Fatalf("db connection failed: %v\n", err)
	}

	if os.Getenv(ep.ENV_DEV) == "true" {
		// TODO: deletion is solely for development to eliminate cumulative state
		ep.DeleteDb(db, cfg.DbCreds.DbName, ep.PostsTable)
	}

	if err = ep.EnsureDB(db, cfg.DbCreds.DbName, &ep.Post{}); err != nil {
		log.Fatalf("%s db creation failed: %v\n", cfg.DbCreds.DbName, err)
	} else {
		log.Printf("%s db exists\n", cfg.DbCreds.DbName)
	}

	log.Printf("Listening at %s\n", cfg.Addr)

	opts := []grpc.ServerOption{}
	gs := grpc.NewServer(opts...)
	ep := ep.NewServer(db)
	pb.RegisterCrudServiceServer(gs, ep)

	if err := gs.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
	}
}
