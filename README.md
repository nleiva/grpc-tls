# gRPC TLS testing

## Run

- Client

`ID` is the user id we want to retrieve.

```bash
ID=1 make run-client
```

- Server

```bash
make run-server
```

## Generating TSL Certificates

Certificates valid for a year (`-days 365`).

- CA Signed certificates

[Example](https://gist.github.com/fntlnz/cf14feb5a46b2eda428e000157447309)

Create Root signing Key

```bash
openssl genrsa -out ca.key 4096
```

Generate self-signed Root certificate

```bash
openssl req -new -x509 -key ca.key -sha256 -subj "/C=US/ST=NJ/O=CA, Inc." -days 365 -out ca.cert
```

Create a Key certificate for your service

```bash
openssl genrsa -out service.key 4096
```

Create signing CSR

For local testing you can use `'/CN=localhost'`. For Online testing `CN` needs to be replaced with your gRPC Server, for example: `'/CN=grpc.nleiva.com'`.

```bash
openssl req -new -sha256 -key service.key -subj "/C=US/ST=NJ/O=Test, Inc./CN=localhost" -out service.csr
```

or

```bash
openssl req -new -key service.key -out service.csr -config certificate.conf
```

Generate a certificate for your service

```bash
openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial -out service.pem -days 365 -sha256
```

or

```bash
openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial -out service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext
```

Verify

```bash
openssl x509 -in service.pem -text -noout
```

## Compiling protocol buffers

```bash
cd proto
protoc --go_out=plugins=grpc:. comm.proto
```

## Running in Docker Containers

Setup your enviroment variables in `env.client` and `env.server`.

- Client

```bash
docker build -t client -f ./client/Dockerfile .
docker run -t --rm --name my-client -e "ID=1" client
```

- Server

```bash
docker build -t server -f ./server/Dockerfile .
docker run -t --rm --name my-server server
```

Just in case...

[docker build](https://docs.docker.com/edge/engine/reference/commandline/build/#usage)

```bash
--tag , -t    Name and optionally a tag in the ‘name:tag’ format
--file , -f   Name of the Dockerfile (Default is ‘PATH/Dockerfile’)
```

[docker run](https://docs.docker.com/edge/engine/reference/commandline/container_run/#usage)

```bash
--tty , -t    Allocate a pseudo-TTY
--rm          Automatically remove the container when it exits
--name        Assign a name to the container
--env-file    Read in a file of environment variables
```
