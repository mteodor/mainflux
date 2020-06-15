// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import "time"

type Cert struct {
	ThingID string
	Serial  string
	Expire  time.Time
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
	//Save()
	Save(cert Cert) (string, error)
}
