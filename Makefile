ID?=1
.DEFAULT_GOAL := help

.EXPORT_ALL_VARIABLES:
PORT?=50051
HOST?=localhost
CAFILE?="ca.cert"

.PHONY: proto

all: cert docker-build

proto: ## Compile protobuf file to generate Go source code for gRPC Services
	protoc --go_out=plugins=grpc:. proto/gumi.proto

cert: ## Create certificates to encrypt the gRPC connection
	openssl genrsa -out ca.key 4096
	openssl req -new -x509 -key ca.key -sha256 -subj "/C=US/ST=NJ/O=CA, Inc." -days 365 -out ca.cert
	openssl genrsa -out service.key 4096
	openssl req -new -key service.key -out service.csr -config certificate.conf
	openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial \
		-out service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext

install-docker: ## Install Docker
	curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
	sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
	sudo apt-get update
	sudo apt-get install -y docker-ce
	sudo usermod -aG docker ${USER}

docker-build: ## Build Docker images for Client and Server
	docker build -t client --build-arg host=${HOST} \
		--build-arg port=${PORT} -f ./client/Dockerfile .
	docker build -t server --build-arg host=${HOST} \
		--build-arg port=${PORT} -f ./server/Dockerfile .

run-docker-client: ## Run Client Docker image with a given ID
	docker run -t --rm --name my-client -e $(ID) client

run-client: ## Run Client with a given ID
	go run client/main.go -id $(ID) -mode 1

run-client-noca: ## Run Client with a given ID without CA
	go run client/main.go -id $(ID) -mode 2

run-client-ca: ## Run Client with a given ID with CA
	go run client/main.go -id $(ID) -file $(CAFILE) -mode 3

run-client-file: ## Run Client with a given ID and provide the Server certificate
	go run client/main.go -id $(ID) -mode 4

run-client-insecure: ## Run Client with no secutity/encryption
	go run client/main.go -mode 5

run-client-default: ## Run Client with default TLS config
	go run client/main.go -mode 6

run-docker-server: ## Run Server Docker image
	docker run -t --rm \
    	--name my-server \
    	--publish=${PORT}:${PORT} \
    	server

run-server: ## Run Server
	go run server/main.go

run-server-insecure: ## Run Server insecurely (no encryption)
	go run server/main.go -self=false

run-server-vault: ## Run Server with certs from Vault
	go run server/main.go -self=false -cefy=true

run-server-public: ## Run Server using Let's Encrypt cert and ACME
	go build -o server/server server/main.go
	sudo setcap CAP_NET_BIND_SERVICE+ep server/server
	server/server -self=false -public=true

docker-stop: ## Stop any Docker images running
	-@docker stop my-server
	-@docker stop my-client

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
