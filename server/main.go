/*
gRPC Server
*/

package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/johanbrandhorst/certify"
	"github.com/johanbrandhorst/certify/issuers/vault"
	pb "github.com/nleiva/grpc-tls/proto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kit "github.com/goph/logur/adapters/kitlogadapter"
)

var (
	host   = getenv("HOST") // E.g.: "test.nleiva.com" or "localhost"
	port   = getenv("PORT") // E.g.: "443" or "50051"
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
)

func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		level.Error(logger).Log("msg", "environment variable not set", "var", name)
		os.Exit(1)
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

type RSA struct {
	bits int
}

func (r RSA) Generate() (crypto.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, r.bits)
}

func vaultCert(f string) (credentials.TransportCredentials, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("vaultCert: problem with input file")
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		return nil, fmt.Errorf("vaultCert: failed to append certificates")
	}
	issuer := &vault.Issuer{
		URL: &url.URL{
			Scheme: "https",
			Host:   "localhost:8200",
		},
		TLSConfig: &tls.Config{
			RootCAs: cp,
		},
		Token: getenv("TOKEN"),
		Role:  "my-role",
	}
	cfg := certify.CertConfig{
		SubjectAlternativeNames: []string{"localhost"},
		IPSubjectAlternativeNames: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
		KeyGenerator: RSA{bits: 2048},
	}
	c := &certify.Certify{
		CommonName:  "localhost",
		Issuer:      issuer,
		Cache:       certify.NewMemCache(),
		CertConfig:  &cfg,
		RenewBefore: 24 * time.Hour,
		Logger:      kit.New(logger),
	}
	tlsConfig := &tls.Config{
		GetCertificate: c.GetCertificate,
	}
	return credentials.NewTLS(tlsConfig), nil
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
	self := flag.Bool("self", true, "Whether to encrypt the connection using self-signed certs")
	public := flag.Bool("public", false, "Use certs emited by a trusted public CA")
	cefy := flag.Bool("cefy", false, "Use Certify with Vault as your CA")
	flag.Parse()

	logger = log.With(logger, "time", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	opts := []grpc.ServerOption{}
	var lis net.Listener
	var err error
	switch {
	// Public domain
	case *public:
		opts = append(opts, grpc.Creds(acmeCert()))
		lis = autocert.NewListener(host)
		port = "443"
	// Private domain
	default:
		switch {
		case *self && *cefy:
			level.Error(logger).Log("msg", "can't choose self-signed and Certify at the same time")
			os.Exit(1)
		// Self-signed cetificates
		case *self:
			creds, err := credentials.NewServerTLSFromFile("service.pem", "service.key")
			if err != nil {
				level.Error(logger).Log("msg", "failed to setup TLS with local files", "error", err)
				os.Exit(1)
			}
			opts = append(opts, grpc.Creds(creds))
		// Certificates signed by Vault via Certify
		case *cefy:
			creds, err := vaultCert("ca-org.cert")
			if err != nil {
				level.Error(logger).Log("msg", "failed to setup TLS with Certify", "error", err)
				os.Exit(1)
			}
			opts = append(opts, grpc.Creds(creds))
		// Insecure
		default:
		}
		lis, err = net.Listen("tcp", ":"+port)
		if err != nil {
			level.Error(logger).Log("msg", "failed to listen", "error", err)
			os.Exit(1)
		}
	}
	defer lis.Close()
	level.Info(logger).Log("msg", "Server listening", "port", port)

	s := grpc.NewServer(opts...)
	level.Info(logger).Log("msg", "Starting gRPC services")
	pb.RegisterGUMIServer(s, NewUserData())

	level.Info(logger).Log("msg", "Listening for incoming connections")
	if *public {
		if err = http.Serve(lis, grpcHandlerFunc(s, httpsHandler())); err != nil {
			level.Error(logger).Log("msg", "failed to serve", "error", err)
			os.Exit(1)
		}
	} else {
		if err = s.Serve(lis); err != nil {
			level.Error(logger).Log("msg", "failed to serve", "error", err)
			os.Exit(1)
		}
	}
}
