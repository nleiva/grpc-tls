# gRPC TLS testing

Basic service to retrive user names based on their `ID`. This is just for TLS testing purposes.

## Run

- Server

```bash
make run-server
```

- Client

`ID` is the user id we want to retrieve.

```bash
export ID=1
make run-client
```

- Help

```bash
make
```

## Generating TSL Certificates

You need these before running the app. To create them run `make cert`. The certificates are valid for a year (`-days 365`).

Below the step by step:

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

or with a config file ([certificate.conf](certificate.conf)):

```bash
openssl req -new -key service.key -out service.csr -config certificate.conf
```

Generate a certificate for your service

```bash
openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial -out service.pem -days 365 -sha256
```

or with a config file ([certificate.conf](certificate.conf)):

```bash
openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial -out service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext
```

Verify

```bash
openssl x509 -in service.pem -text -noout
```

## Running in Docker Containers

Build Docker images with `make docker-build`.

- Run the Docker Client image. Provide any `ID`.

```bash
export ID=1
make run-docker-client
```

- Run the Docker Server image

```bash
make run-docker-server
```

## Compiling protocol buffers

Run `make proto`.
