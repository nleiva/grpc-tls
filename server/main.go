/*
gRPC Server
*/

package main

import (
	"context"
	"log"
	"net"

	pb "github.com/nleiva/grpc-tls/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port = ":50051"
)

type server struct {
	users map[uint32]string
}

func (s *server) GetByID(ctx context.Context, in *pb.GetByIDRequest) (*pb.User, error) {
	if s.users == nil {
		s.users = make(map[uint32]string)
	}
	if name, ok := s.users[in.Id]; ok {
		return &pb.User{
			Name: name,
			Id:   in.Id,
		}, nil
	}
	return nil, errors.New("user not found")
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Security options
	creds, err := credentials.NewServerTLSFromFile("service.pem", "service.key")
	if err != nil {
		log.Fatalf("Failed to setup tls: %v", err)
	}
	opts := []grpc.ServerOption{grpc.Creds(creds)}
	// Setup a secure Server
	s := grpc.NewServer(opts...)

	srv := new(server)
	srv.users = make(map[uint32]string)
	srv.users[1] = "Nicolas"

	pb.RegisterGUMIServer(s, srv)
	log.Println("Starting server on port " + port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
