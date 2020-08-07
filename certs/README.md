# Certs Service
Issues certificates for things. `Certs` service can create certificates to be used when `Mainflux` is deployed to support mTLS.
Certificate service can create certificates in two modes:
1. Development mode - to be used when no PKI is deployed, this works similar to the [make thing_cert](../docker/ssl/Makefile)
2. PKI mode - certificates issued by PKI, when you deploy `Vault` as PKI certificate management `cert` service will proxy requests to `Vault` previously checking access rights and saving info on successfully created certificate. 



## Configuration

The service is configured using the environment variables presented in the following table. Note that any unset variables will be replaced with their default values.

| Variable                      | Description                                                             | Default                          |
|-------------------------------|-------------------------------------------------------------------------|----------------------------------|
| MF_CERTS_LOG_LEVEL            | Log level for Certs (debug, info, warn, error)                          | error                            |
| MF_CERTS_DB_HOST              | Database host address                                                   | localhost                        |
| MF_CERTS_DB_PORT              | Database host port                                                      | 5432                             |
| MF_CERTS_DB_USER              | Database user                                                           | mainflux                         |
| MF_CERTS_DB_PASS              | Database password                                                       | mainflux                         |
| MF_CERTS_DB                   | Name of the database used by the service                                | certs                            |
| MF_CERTS_DB_SSL_MODE          | Database connection SSL mode (disable, require, verify-ca, verify-full) | disable                          |
| MF_CERTS_DB_SSL_CERT          | Path to the PEM encoded certificate file                                |                                  |
| MF_CERTS_DB_SSL_KEY           | Path to the PEM encoded key file                                        |                                  |
| MF_CERTS_DB_SSL_ROOT_CERT     | Path to the PEM encoded root certificate file                           |                                  |
| MF_CERTS_CLIENT_TLS           | Flag that indicates if TLS should be turned on                          | false                            |
| MF_CERTS_CA_CERTS             | Path to trusted CAs in PEM format                                       |                                  |
| MF_CERTS_PORT                 | Certs service HTTP port                                                 | 8204                             |
| MF_CERTS_SERVER_CERT          | Path to server certificate in pem format                                |                                  |
| MF_CERTS_SERVER_KEY           | Path to server key in pem format                                        |                                  |
| MF_SDK_BASE_URL               | Base url for Mainflux SDK                                               | http://localhost                 |
| MF_SDK_THINGS_PREFIX          | SDK prefix for Things service                                           |                                  |
| MF_THINGS_ES_URL              | Things service event source URL                                         | localhost:6379                   |
| MF_THINGS_ES_PASS             | Things service event source password                                    |                                  |
| MF_THINGS_ES_DB               | Things service event source database                                    | 0                                |
| MF_JAEGER_URL                 | Jaeger server URL                                                       | localhost:6831                   |
| MF_AUTHN_GRPC_URL             | AuthN service gRPC URL                                                  | localhost:8181                   |
| MF_AUTHN_GRPC_TIMEOUT         | AuthN service gRPC request timeout in seconds                           | 1s                               |
| MF_CERTS_SIGN_CA_PATH         | CA certificate for signing in Development mode                          | "ca.crt"                         |
| MF_CERTS_SIGN_CA_KEY_PATH     | CA certificate signing key for signing in Development mode              | "ca.key"                         |
| MF_CERTS_SIGN_HOURS_VALID     | Default certificate valid period                                        | 2048h                            |
| MF_CERTS_SIGN_RSA_BITS        | Default number of RSA bits for certificate                              | 2048                             |
| MF_CERTS_VAULT_HOST           | Vault host address, if not set Development mode is used                 | ""                               |
| MF_CERTS_VAULT_PKI_PATH       | Vault PKI path, path where certificates are issued                      | pki_int                          |
| MF_CERTS_VAULT_ROLE           | Vault PKI role that is used for issuing certificate                     | mainflux                         |
| MF_CERTS_VAULT_TOKEN          | Vault API access token                                                  | ""                               |


## Development mode
If `MF_CERTS_VAULT_HOST` is empty than Development mode is on.

To issue a certificate:
```bash

TOK=`curl  -s --insecure -S -X POST http://localhost/tokens -H 'Content-Type: application/json' -d '{"email":"edge@email.com","password":"12345678"}' | jq -r '.token'`

curl -s -S  -X POST  http://localhost:8204/certs -H "Authorization: $TOK" -H 'Content-Type: application/json'   -d '{"thing_id":<thing_id>, "rsa_bits":2048, "key_type":"rsa"}'
```

```json
{
  "ThingID": "",
  "ClientCert": "-----BEGIN CERTIFICATE-----\nMIIDmTCCAoGgAwIBAgIRANmkAPbTR1UYeYO0Id/4+8gwDQYJKoZIhvcNAQELBQAw\nVzESMBAGA1UEAwwJbG9jYWxob3N0MREwDwYDVQQKDAhNYWluZmx1eDEMMAoGA1UE\nCwwDSW9UMSAwHgYJKoZIhvcNAQkBFhFpbmZvQG1haW5mbHV4LmNvbTAeFw0yMDA2\nMzAxNDIxMDlaFw0yMDA5MjMyMjIxMDlaMFUxETAPBgNVBAoTCE1haW5mbHV4MREw\nDwYDVQQLEwhtYWluZmx1eDEtMCsGA1UEAxMkYjAwZDBhNzktYjQ2YS00NTk3LTli\nNGYtMjhkZGJhNTBjYTYyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\ntgS2fLUWG3CCQz/l6VRQRJfRvWmdxK0mW6zIXGeeOILYZeaLiuiUnohwMJ4RiMqT\nuJbInAIuO/Tt5osfrCFFzPEOLYJ5nZBBaJfTIAxqf84Ou1oeMRll4wpzgeKx0rJO\nXMAARwn1bT9n3uky5QQGSLy4PyyILzSXH/1yCQQctdQB/Ar/UI1TaYoYlGzh7dHT\nWpcxq1HYgCyAtcrQrGD0rEwUn82UBCrnya+bygNqu0oDzIFQwa1G8jxSgXk0mFS1\nWrk7rBipsvp8HQhdnvbEVz4k4AAKcQxesH4DkRx/EXmU2UvN3XysvcJ2bL+UzMNI\njNhAe0pgPbB82F6zkYZ/XQIDAQABo2IwYDAOBgNVHQ8BAf8EBAMCB4AwHQYDVR0l\nBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMBMA4GA1UdDgQHBAUBAgMEBjAfBgNVHSME\nGDAWgBRs4xR91qEjNRGmw391xS7x6Tc+8jANBgkqhkiG9w0BAQsFAAOCAQEAW/dS\nV4vNLTZwBnPVHUX35pRFxPKvscY+vnnpgyDtITgZHYe0KL+Bs3IHuywtqaezU5x1\nkZo+frE1OcpRvp7HJtDiT06yz+18qOYZMappCWCeAFWtZkMhlvnm3TqTkgui6Xgl\nGj5xnPb15AOlsDE2dkv5S6kEwJGHdVX6AOWfB4ubUq5S9e4ABYzXGUty6Hw/ZUmJ\nhCTRVJ7cQJVTJsl1o7CYT8JBvUUG75LirtoFE4M4JwsfsKZXzrQffTf1ynqI3dN/\nHWySEbvTSWcRcA3MSmOTxGt5/zwCglHDlWPKMrXtjTW7NPuGL5/P9HSB9HGVVeET\nDUMdvYwgj0cUCEu3LA==\n-----END CERTIFICATE-----\n",
  "IssuingCA": "",
  "CAChain": null,
  "ClientKey": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAtgS2fLUWG3CCQz/l6VRQRJfRvWmdxK0mW6zIXGeeOILYZeaL\niuiUnohwMJ4RiMqTuJbInAIuO/Tt5osfrCFFzPEOLYJ5nZBBaJfTIAxqf84Ou1oe\nMRll4wpzgeKx0rJOXMAARwn1bT9n3uky5QQGSLy4PyyILzSXH/1yCQQctdQB/Ar/\nUI1TaYoYlGzh7dHTWpcxq1HYgCyAtcrQrGD0rEwUn82UBCrnya+bygNqu0oDzIFQ\nwa1G8jxSgXk0mFS1Wrk7rBipsvp8HQhdnvbEVz4k4AAKcQxesH4DkRx/EXmU2UvN\n3XysvcJ2bL+UzMNIjNhAe0pgPbB82F6zkYZ/XQIDAQABAoIBAALoal3tqq+/iWU3\npR2oKiweXMxw3oNg3McEKKNJSH7QoFJob3xFoPIzbc9pBxCvY9LEHepYIpL0o8RW\nHqhqU6olg7t4ZSb+Qf1Ax6+wYxctnJCjrO3N4RHSfevqSjr6fEQBEUARSal4JNmr\n0hNUkCEjWrIvrPFMHsn1C5hXR3okJQpGsad4oCGZDp2eZ/NDyvmLBLci9/5CJdRv\n6roOF5ShWweKcz1+pfy666Q8RiUI7H1zXjPaL4yqkv8eg/WPOO0dYF2Ri2Grk9OY\n1qTM0W1vi9zfncinZ0DpgtwMTFQezGwhUyJHSYHmjVBA4AaYIyOQAI/2dl5fXM+O\n9JfXpOUCgYEA10xAtMc/8KOLbHCprpc4pbtOqfchq/M04qPKxQNAjqvLodrWZZgF\nexa+B3eWWn5MxmQMx18AjBCPwbNDK8Rkd9VqzdWempaSblgZ7y1a0rRNTXzN5DFP\noiuRQV4wszCuj5XSdPn+lxApaI/4+TQ0oweIZCpGW39XKePPoB5WZiMCgYEA2G3W\niJncRpmxWwrRPi1W26E9tWOT5s9wYgXWMc+PAVUd/qdDRuMBHpu861Qoghp/MJog\nBYqt2rQqU0OxvIXlXPrXPHXrCLOFwybRCBVREZrg4BZNnjyDTLOu9C+0M3J9ImCh\n3vniYqb7S0gRmoDM0R3Zu4+ajfP2QOGLXw1qHH8CgYEAl0EQ7HBW8V5UYzi7XNcM\nixKOb0YZt83DR74+hC6GujTjeLBfkzw8DX+qvWA8lxLIKVC80YxivAQemryv4h21\nX6Llx/nd1UkXUsI+ZhP9DK5y6I9XroseIRZuk/fyStFWsbVWB6xiOgq2rKkJBzqw\nCCEQpx40E6/gsqNDiIAHvvUCgYBkkjXc6FJ55DWMLuyozfzMtpKsVYeG++InSrsM\nDn1PizQS/7q9mAMPLCOP312rh5CPDy/OI3FCbfI1GwHerwG0QUP/bnQ3aOTBmKoN\n7YnsemIA/5w16bzBycWE5x3/wjXv4aOWr9vJJ/siMm0rtKp4ijyBcevKBxHpeGWB\nWAR1FQKBgGIqAxGnBpip9E24gH894BaGHHMpQCwAxARev6sHKUy27eFUd6ipoTva\n4Wv36iz3gxU4R5B0gyfnxBNiUab/z90cb5+6+FYO13kqjxRRZWffohk5nHlmFN9K\nea7KQHTfTdRhOLUzW2yVqLi9pzfTfA6Yqf3U1YD3bgnWrp1VQnjo\n-----END RSA PRIVATE KEY-----\n",
  "PrivateKeyType": "",
  "Serial": "",
  "Expire": "0001-01-01T00:00:00Z"
}
```

## PKI mode

When `MF_CERTS_VAULT_HOST` is set it is presumed that `Vault` is installed and `certs` service will issue certificates using `Vault` API.
First you'll need to set up `Vault`. 
To setup `Vault` follow steps in [Build Your Own Certificate Authority (CA)](https://learn.hashicorp.com/tutorials/vault/pki-engine).

To setup certs service with `Vault` following environment variables must be set:

```
MF_CERTS_VAULT_HOST=vault-domain.com
MF_CERTS_VAULT_PKI_PATH=<vault_pki_path>
MF_CERTS_VAULT_ROLE=<vault_role>
MF_CERTS_VAULT_TOKEN=<vault_acces_token>
```

For lab purposes you can use docker-compose and script for setting up PKI in [https://github.com/mteodor/vault](https://github.com/mteodor/vault)

Issuing certificate is same as in **Development** mode.
In this mode certificates can also be revoked:

```bash
curl -s -S -X DELETE http://localhost:8204/certs/revoke -H "Authorization: $TOK" -H 'Content-Type: application/json'   -d '{"thing_id":"c30b8842-507c-4bcd-973c-74008cef3be5"}'
```