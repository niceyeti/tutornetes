package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "go_grpc_example/proto"
)

var addr string = "127.0.0.1:80"

func createPost(c pb.CrudServiceClient) {
	log.Println("createPost was invoked")

	res, err := c.CreatePost(context.Background(), &pb.Post{
		Id:          "321123",
		AuthorId:    "Jose",
		Title:       "Gone With the Wind",
		Description: "Humpty dumpy",
		FullText:    "In the beginning...",
	})
	if err != nil {
		e, ok := status.FromError(err)
		if ok {
			log.Printf("Error message from server: %v\n", e.Message())
			log.Println("Code: ", e.Code())

			if e.Code() == codes.InvalidArgument {
				log.Println("We probably sent a negative number!")
			}
		} else {
			log.Fatalf("A non gRPC error: %v\n", err)
		}
		return
	}

	log.Printf("CreatePost response: %v\n", res)
}

func main() {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v\n", err)
	}

	defer conn.Close()
	c := pb.NewCrudServiceClient(conn)

	createPost(c)
}
