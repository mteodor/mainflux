// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import "time"

type Cert struct {
	ThingID      string
	ThingKey     string
	ClientCert   string
	ClientKey    string
	CACert       string
	ClientCertID string
	Expire       time.Time
}

// ConfigsPage contains page related metadata as well as list
type CertsPage struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Certs  []Cert
}

// CertsRepository specifies a Config persistence API.
type CertsRepository interface {
	//AddCert()
	AddCert()
}
