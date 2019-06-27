/*
gRPC Client
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "github.com/nleiva/grpc-tls/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	host = "localhost:50051"
)

func main() {
	// YANG path arguments; defaults to "yangocpaths.json"
	id := flag.Uint("id", 1, "User ID")
	flag.Parse()
	// Security options
	/* config := &tls.Config{
		InsecureSkipVerify: true,
	} */
	ctx := context.Background()
	//opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(config))}
	opts := []grpc.DialOption{}
	creds, err := credentials.NewClientTLSFromFile("service.pem", "")
	if err != nil {
		log.Fatalf("could not process the credentials: %v", err)
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	// Set up a secure connection to the server.
	conn, err := grpc.Dial(host, opts...)
	//conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGUMIClient(conn)
	res, err := client.GetByID(ctx, &pb.GetByIDRequest{Id: uint32(*id)})
	if err != nil {
		log.Fatalf("Server says: %v", err)
	}
	fmt.Println("User found: ", res.GetName())
}
