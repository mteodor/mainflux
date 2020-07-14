// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package postgres contains repository implementations using PostgreSQL as
// the underlying database.
package pki

import "time"

type Revoke struct {
	RevocationTime time.Time `mapstructure:"revocation_time"`
}

type Cert struct {
	ClientCert     string    `json:"client_cert" mapstructure:"certificate"`
	IssuingCA      string    `json:"issuing_ca" mapstructure:"issuing_ca"`
	CAChain        []string  `json:"ca_chain" mapstructure:"ca_chain"`
	ClientKey      string    `json:"client_key" mapstructure:"private_key"`
	PrivateKeyType string    `json:"private_key_type" mapstructure:"private_key_type"`
	Serial         string    `json:"serial" mapstructure:"serial_number"`
	Expire         time.Time `json:"expire" mapstructure:"-"`
}

type Agent interface {
	// IssueCert issues certificate on PKI
	IssueCert(cn string, ttl, keyType string, keyBits int) (Cert, error)
	// Revoke revokes certificate from PKI
	Revoke(serial string) (Revoke, error)
}
