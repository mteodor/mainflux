package certs

import (
	"github.com/mainflux/mainflux/certs/vault"
)

type PKI interface {
	// IssueCert issues certificate on PKI
	IssueCert(cn string, ttl, keyType string, keyBits int) (vault.Cert, error)
	// Revoke revokes certificate from PKI
	Revoke(serial string) (vault.Revoke, error)
}
