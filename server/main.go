/*
gRPC Server
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	pb "github.com/nleiva/grpc-tls/proto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	host = getenv("HOST") // Eg: "test.nleiva.com"
	port = getenv("PORT")
)

func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Panicf("%s environment variable not set.", name)
	}
	return v
}

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

func acmeCert() credentials.TransportCredentials {
	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("golang-autocert"),
		HostPolicy: autocert.HostWhitelist(host),
		// Email: "",
	}
	return credentials.NewTLS(manager.TLSConfig())
}

func grpcHandlerFunc(g *grpc.Server, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if r.ProtoMajor == 2 && strings.Contains(ct, "application/grpc") {
			g.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func httpsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS user from IP: %s\n\nYour config is: %+v", r.RemoteAddr, r.TLS)
	})
}

func main() {
	secure := flag.Bool("secure", true, "Whether to encrypt the connection using self-signed certs")
	public := flag.Bool("public", false, "Use certs emited by a trusted CA")
	flag.Parse()

	opts := []grpc.ServerOption{}
	var lis net.Listener
	var err error
	if *secure {
		creds, err := credentials.NewServerTLSFromFile("service.pem", "service.key")
		if err != nil {
			log.Fatalf("Failed to setup TLS: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
		lis, err = net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
	}
	if *public {
		opts = append(opts, grpc.Creds(acmeCert()))
		lis = autocert.NewListener(host)
		port = "443"
	}
	defer lis.Close()
	log.Println("Creating listener on port:", port)

	s := grpc.NewServer(opts...)
	pb.RegisterGUMIServer(s, NewUserData())
	log.Println("Starting gRPC services")

	log.Println("Listening for incomming connections")
	if *public {
		if err = http.Serve(lis, grpcHandlerFunc(s, httpsHandler())); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	} else {
		if err = s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}
}
