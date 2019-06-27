ID?=1

all: cert docker-build

pb:
	cd proto
	protoc --go_out=plugins=grpc:. comm.proto

cert:
	openssl genrsa -out ca.key 4096
	openssl req -new -x509 -key ca.key -sha256 -subj "/C=US/ST=NJ/O=CA, Inc." -days 365 -out ca.cert
	openssl genrsa -out service.key 4096
	openssl req -new -key service.key -out service.csr -config certificate.conf
	openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial \
		-out service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext

docker-build: 
	docker build -t client -f ./client/Dockerfile .
	docker build -t server -f ./server/Dockerfile .

run-docker-client:
	docker run -t --rm --name my-client -e $(ID) client

run-client:
	go run client/main.go -id $(ID)

run-docker-server:
	docker run -t --rm --name my-server server

run-server:
	go run server/main.go

docker-stop:
	docker stop my-server
	docker stop my-client