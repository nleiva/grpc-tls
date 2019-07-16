# gRPC TLS testing

Basic service to retrive user names based on their `ID`. This is just for TLS testing purposes.

## Run

- Server

    ```bash
    make run-server
    ```

- Client

You need to provide an `ID` which is the id of the user we want to retrieve from the Server, for example `export ID=1`.

1. Connect using the cert the Server provides during the TLS Handshake without verifying it.

    ```bash
    make run-client
    ```

2. Connect using the cert the Server provides during the TLS Handshake and verify it.

    ```bash
    make run-client-noca
    ```

3. Connect using the cert the Server provides during the TLS Handshake and verify it with a CA cert file provided.

    ```bash
    make run-client-ca
    ```

4. Connect using a cert provided at runtime.

    ```bash
    make run-client-file
    ```

- Help

    ```bash
    make
    ```

## Generating TSL Certificates

You need these before running the examples. To create them run `make cert`. The certificates are valid for a year (`-days 365`). Below the step by step, for your reference.

- CA Signed certificates

1. Create Root signing Key

    ```bash
    openssl genrsa -out ca.key 4096
    ```

2. Generate self-signed Root certificate

    ```bash
    openssl req -new -x509 -key ca.key -sha256 -subj "/C=US/ST=NJ/O=CA, Inc." -days 365 -out ca.cert
    ```

3. Create a Key certificate for your service

    ```bash
    openssl genrsa -out service.key 4096
    ```

4. Create signing CSR

    For local testing you can use `'/CN=localhost'`. For Online testing `CN` needs to be replaced with your gRPC Server, for example: `'/CN=grpc.nleiva.com'`. Include this in a config file ([certificate.conf](certificate.conf)).

    ```bash
    openssl req -new -key service.key -out service.csr -config certificate.conf
    ```

5. Generate a certificate for the service

    ```bash
    openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial -out service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext
    ```

6. Verify

    ```bash
    openssl x509 -in service.pem -text -noout
    ```

## Vault and Certify

See [vault-cert.md](vault-cert.md) for setup details.

- Server

    ```bash
    make run-server-vault
    ```

- Client

    ```bash
    export CAFILE="ca-vault.cert"
    make run-client-ca
    ```

You need to provide an `ID` which is the id of the user we want to retrieve from the Server, for example `export ID=1`. Also, the name of the Vault's CA certificate file as `CAFILE`.

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
