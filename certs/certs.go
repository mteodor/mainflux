// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import "context"

// ConfigsPage contains page related metadata as well as list
type Page struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Certs  []Cert
}

// Repository specifies a Config persistence API.
type Repository interface {
	// Save  saves cert for thing into database
	Save(ctx context.Context, cert Cert) (string, error)

	// RetrieveAll retrieve all issued certificates for given owner
	RetrieveAll(ctx context.Context, ownerID string, offset, limit uint64) (Page, error)

	// Retrieve retrieve issued certificates for given thing
	Retrieve(ctx context.Context, cert Cert, offset, limit uint64) (Page, error)

	// Remove certificate from DB
	Remove(ctx context.Context, cert Cert) error
}
