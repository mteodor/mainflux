// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import "context"

// ConfigsPage contains page related metadata as well as list
type CertsPage struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Certs  []Cert
}

// CertsRepository specifies a Config persistence API.
type CertsRepository interface {
	//Save  saves cert for thing into database
	Save(ctx context.Context, cert Cert) (string, error)

	//RetrieveAll retrieve all issued certificates for given thing
	RetrieveAll(ctx context.Context, thingID string, offset, limit uint64) (CertsPage, error)
}
