ui = true
default_lease_ttl = "168h"
max_lease_ttl = "720h"
disable_mlock = true

storage "file" {
  path = "/Users/nleiva/vault/data"
}

listener "tcp" {
  address     = "localhost:8200"
  tls_cert_file = "/Users/nleiva/vault/vault.pem"
  tls_key_file = "/Users/nleiva/vault/vault.key"
}