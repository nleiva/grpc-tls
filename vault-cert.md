# Setting up Vault and Certify

## Vault

1. Vault comes as a binary you can place on any path in your `$PATH`. . Follow instructions from [Vault](https://www.vaultproject.io/downloads.html) website.

2. Start the server.

```bash
vault server -config=vault_config.hcl
```

3. Initialize the server.

ca.cert: "This file is used to verify the Vault server's SSL certificate" [VAULT](https://www.vaultproject.io/docs/commands/index.html#vault_addr)

```bash
export VAULT_ADDR=https://localhost:8200
export VAULT_CACERT=ca.cert
vault operator init
```

4. Unseal the Vault.

```bash
export uKey1="..."
export uKey2="..."
export uKey3="..."
vault operator unseal ${uKey1}
vault operator unseal ${uKey2}
vault operator unseal ${uKey3}
```

5. Test Vault.

```bash
$ curl \
    --cacert ca.cert \
    -i https://localhost:8200/v1/sys/health
HTTP/1.1 200 OK
Cache-Control: no-store
Content-Type: application/json
Date: Mon, 15 Jul 2019 01:42:29 GMT
Content-Length: 294

{"initialized":true,"sealed":false,"standby":false,"performance_standby":false,"replication_performance_mode":"disabled","replication_dr_mode":"disabled","server_time_utc":1563154949,"version":"1.1.3","cluster_name":"vault-cluster-d6f1a7ef","cluster_id":"50b7cade-fd03-c05c-9b19-05467bd285e7"}
```

6. Enable Vault PKI Secrets Engine backend. Follow instructions from [Vault](https://www.vaultproject.io/docs/secrets/pki/index.html) website.

```bash
vault secrets enable pki
```

Generate CA certificate and private key.

```bash
vault write pki/root/generate/internal \
    common_name=localhost \
    ttl=8760h
```

Certificate location.

```bash
vault write pki/config/urls \
    issuing_certificates="https://localhost:8200/v1/pki/ca" \
    crl_distribution_points="https://localhost:8200/v1/pki/crl"
```

Create a role (`my-role`).

```bash
vault write pki/roles/my-role \
    allowed_domains=localhost \
    allow_subdomains=true \
    max_ttl=72h
```

7. Test Vault PKI (with `vault_cn.json`).

```json
{
  "common_name": "localhost"
}
```

```bash
export TOKEN="..."
curl \
    --cacert ca.cert \
    --header "X-Vault-Token: ${TOKEN}" \
    --request POST \
    --data @vault_cn.json \
    https://localhost:8200/v1/pki/issue/my-role
{"request_id":"b1248933-c113-291d-61ac-59487ff8c27c","lease_id":"","renewable":false,"lease_duration":0,"data":{"certificate":"-----BEGIN CERTIFICATE-----\nMIIDsTCCApmgAwIBAgIUKyqoOpSkEgptLE3LOyrn/oE1MoUwDQYJKoZIhvcNAQEL\nBQAwFDESMB...-----END CERTIFICATE-----", ... }
```

## Certify

- Run the server with `make run-server-vault`

```bash
$ make run-server-vault
go run server/main.go -secure=false -cefy=true
2019/07/14 22:15:12 Creating listener on port: 50051
2019/07/14 22:15:12 Starting gRPC services
2019/07/14 22:15:12 Listening for incoming connections
```

- Run the client with `make run-client`

```bash
$ make run-client
go run client/main.go -id 1 -mode 1
2019/07/14 22:15:43 Server says: rpc error: code = Unavailable desc = all SubConns are in TransientFailure, latest connection error: connection error: desc = "transport: authentication handshake failed: remote error: tls: internal error"
exit status 1
make: *** [run-client] Error 1
```

Vault doesn't create a new certificate (as opposed to what happens when we test it with `curl`)

### Troubleshooting

I see after running the client that we do contact Vault. See [Packet Capture](grpc-certify.pcap).

Just to double-check, when I comment out the CA cert in the issuer.

```go
issuer := &vault.Issuer{
    ...
	//TLSConfig: &tls.Config{
	//	RootCAs: cp,
	//},
	...
	Role:  "my-role",
}
```

Vault, as expected, logs:

```bash
2019-07-12T17:01:10.274-0400 [INFO]  http: TLS handshake error from 127.0.0.1:65080: remote error: tls: bad certificate
2019-07-12T17:01:11.574-0400 [INFO]  http: TLS handshake error from 127.0.0.1:65109: remote error: tls: bad certificate
2019-07-12T17:01:14.346-0400 [INFO]  http: TLS handshake error from 127.0.0.1:65152: remote error: tls: bad certificate
```
