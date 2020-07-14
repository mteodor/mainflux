package certs

import (
	"github.com/mainflux/mainflux/certs/vault"
)

type PKI interface {
	IssueCert(cn string, ttl, keyType string, keyBits int) (vault.Cert, error)
	Revoke(serial string) (vault.Revoke, error)
}
