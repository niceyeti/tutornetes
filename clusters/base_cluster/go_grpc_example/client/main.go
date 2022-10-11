// The Client implementation here is purely for manual testing and development.
// It merely exercises the CRUD api of the gRPC Post service.

package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "go_grpc_example/proto"
)

var addr string = "127.0.0.1:80"

func readPost(c pb.CrudServiceClient, postId *pb.PostID) *pb.Post {
	log.Println("readPost was invoked")

	res, err := c.ReadPost(context.Background(), postId)
	logErr(err)

	log.Printf("ReadPost response: %v\n", res)
	return res
}

func deletePost(c pb.CrudServiceClient, postId *pb.PostID) bool {
	_, err := c.DeletePost(context.Background(), postId)
	logErr(err)
	return err == nil
}

func updatePost(c pb.CrudServiceClient, post *pb.Post) bool {
	_, err := c.UpdatePost(context.Background(), post)
	logErr(err)
	return err == nil
}

func createPost(c pb.CrudServiceClient) *pb.PostID {
	log.Println("createPost was invoked")

	res, err := c.CreatePost(context.Background(), &pb.Post{
		Id:          "321123",
		AuthorId:    "Jose",
		Title:       "Gone With the Wind",
		Description: "Humpty dumpy",
		FullText:    "In the beginning...",
	})
	logErr(err)

	log.Printf("CreatePost response: %v\n", res)
	return res
}

func logErr(err error) {
	if err == nil {
		return
	}

	e, ok := status.FromError(err)
	if ok {
		log.Printf("Error message from server: %v\n", e.Message())
		log.Println("Code: ", e.Code())
		log.Println("Error: ", e.String())
		if e.Code() == codes.InvalidArgument {
			log.Println("We probably sent a negative number!")
		}
	} else {
		log.Printf("A non gRPC error: %v\n", err)
	}
	log.Fatalf("client dying a cowardly end :(")
}

func main() {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v\n", err)
	}

	defer conn.Close()
	cli := pb.NewCrudServiceClient(conn)

	postId := createPost(cli)
	post := readPost(cli, postId)

	post.Description = post.Description + " " + time.Now().Format(time.RFC3339)
	if updatePost(cli, post) {
		log.Println("post updated successfully")
	} else {
		log.Println("post update FAILED")
	}

	_ = readPost(cli, postId)

	if deletePost(cli, postId) {
		log.Printf("deletion of %s succeeded\n", postId)
	} else {
		log.Printf("deletion of %s FAILED\n", postId)
		return
	}

	// This is expected to fail, if the post was successfully deleted
	_ = readPost(cli, postId)
}
