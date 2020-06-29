# CERTS Service
Issues certificates for things. `Certs` service can create certificates to be used when `Mainflux` is deployed to support mTLS.
Certificate service can create certificates in two modes:
1. Development mode - to be used when no PKI is deployed, this works similar to the [make thing_cert](../docker/ssl/Makefile)
2. PKI mode - certificates issued by PKI, when you deploy `Vault` as PKI certificate management `cert` service will proxy requests to `Vault` previously checking access rights and saving info on successfully created certificate. 
   
To issue a certificate:
```bash
curl -s -S  -X POST  http://localhost:8204/certs -H "Authorization: $TOK" -H 'Content-Type: application/json'   -d '{"thing_id":<thing_id>, "rsa_bits":2048, "key_type":"rsa"}'
```