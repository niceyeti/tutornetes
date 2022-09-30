//go:build !test
// +build !test

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	// Note: there is little to no reason to use viper in this app, I just wanted to play with it.
	// It does have terrific functions for reading and monitoring configs locally and remotely
	// for interesting use-cases like hot reloads; none of that is needed in this app.
	"github.com/spf13/viper"

	//pb "github.com/niceyeti/tutornetes/clusters/base_cluster/go_grpc_example/src/proto"
	pb "go_grpc_example/src/proto"
	//pb "github.com/Clement-Jean/grpc-go-course/blog/proto"
	"google.golang.org/grpc"
)

const (
	ENV_SERV_HOST     = "HOST"
	ENV_SERV_PORT     = "PORT"
	SERV_HOST_DEFAULT = "127.0.0.1"
	SERV_PORT_DEFAULT = "80"
	HTTPS_CERT_PATH   = "/etc/secrets/host.cert"
	HTTPS_KEY_PATH    = "/etc/secrets/host.key"

	DB_HOST         = "DB_HOST"
	DB_PORT         = "DB_PORT"
	DB_HOST_DEFAULT = "127.0.0.1"
	DB_PORT_DEFAULT = "5432"
	DB_USER_PATH    = "/etc/secrets/db/user"
	DB_PASS_PATH    = "/etc/secrets/db/passwd"
)

type DBCreds struct {
	Addr string
	User string
	Pass string
}

type AppConfig struct {
	DbCreds DBCreds
	Addr    string
	Cert    string
	Key     string
}

func getEnv(envVar, defaultVal string) string {
	viper.BindEnv(envVar)
	viper.SetDefault(envVar, defaultVal)
	return viper.GetString(envVar)
}

func getTrimmedConfig(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unable to read config file %s: %w", path, err)
	}

	return strings.TrimSpace(string(bytes)), nil
}

func readDBConfig() (*DBCreds, error) {
	dbHost := getEnv(DB_HOST, DB_HOST_DEFAULT)
	dbPort := getEnv(DB_PORT, DB_PORT_DEFAULT)

	dbUser, err := getTrimmedConfig(DB_USER_PATH)
	if err != nil {
		return nil, err
	}

	dbPass, err := getTrimmedConfig(DB_PASS_PATH)
	if err != nil {
		return nil, err
	}

	return &DBCreds{
		Addr: fmt.Sprintf("%s:%s", dbHost, dbPort),
		User: dbUser,
		Pass: dbPass,
	}, nil
}

func readAppConfig() (*AppConfig, error) {
	dbCreds, err := readDBConfig()
	if err != nil {
		return nil, err
	}

	host := getEnv(ENV_SERV_HOST, SERV_HOST_DEFAULT)
	port := getEnv(ENV_SERV_PORT, SERV_PORT_DEFAULT)
	addr := fmt.Sprintf("%s:%s", host, port)

	// TODO: add encryption later. The mesh takes care of this, but it would be a useful exercise.
	//cert := getEnv(HTTPS_CERT_PATH)
	//key := getEnv(HTTPS_KEY_PATH)

	return &AppConfig{
		DbCreds: *dbCreds,
		Addr:    addr,
		Cert:    "",
		Key:     "",
	}, nil
}

func main() {
	//client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:root@localhost:27017/"))
	//if err != nil {
	//	log.Fatal(err)
	//}

	//err = client.Connect(context.Background())
	//if err != nil {
	//	log.Fatal(err)
	//}

	//collection = client.Database("blogdb").Collection("blog")

	cfg, err := readAppConfig()
	if err != nil {
		log.Fatalf("error reading config: %w", err)
	}

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	log.Printf("Listening at %s\n", cfg.Addr)

	s := grpc.NewServer()
	pb.RegisterCrudServiceServer(s, &Server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
	}
}
