/*
gRPC Server
*/

package main

import (
	"context"
	"flag"
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

type userData struct {
	users map[uint32]string
}

func NewUserData() *userData {
	d := new(userData)
	d.users = make(map[uint32]string)
	d.users[1] = "Nicolas"
	return d
}

func (s *userData) GetByID(ctx context.Context, in *pb.GetByIDRequest) (*pb.User, error) {
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
	secure := flag.Bool("secure", true, "Whether to encryt the connection")
	flag.Parse()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Security options
	creds, err := credentials.NewServerTLSFromFile("service.pem", "service.key")
	if err != nil {
		log.Fatalf("Failed to setup tls: %v", err)
	}
	opts := []grpc.ServerOption{}
	if *secure {
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	s := grpc.NewServer(opts...)

	data := NewUserData()

	pb.RegisterGUMIServer(s, data)
	log.Println("Starting server on port " + port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
