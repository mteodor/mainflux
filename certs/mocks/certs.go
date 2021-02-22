// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/logger"
)

var (
	errRemove = errors.New("failed to remove certificate from database")
)

var _ certs.Repository = (*certRepositoryMock)(nil)

type certRepositoryMock struct {
	mu           sync.Mutex
	certs        map[string]certs.Cert
	certsByThing map[string]certs.Cert
	counter      int
}

// NewRepository creates in-memory certs repository.
func NewRepository(log logger.Logger) certs.Repository {
	repo := &certRepositoryMock{
		certs:        make(map[string]certs.Cert),
		certsByThing: make(map[string]certs.Cert),
	}

	return repo
}

func (crm *certRepositoryMock) RetrieveAll(ctx context.Context, ownerID string, offset, limit uint64) (certs.Page, error) {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	if limit <= 0 {
		return certs.Page{}, nil
	}

	first := uint64(offset) + 1
	last := first + uint64(limit)

	var crts []certs.Cert
	idx := uint64(0)
	for _, v := range crm.certs {
		if idx >= first && idx < last {
			crts = append(crts, v)
		}
		idx = idx + 1
	}

	page := certs.Page{
		Certs:  crts,
		Offset: offset,
		Limit:  limit,
	}

	return page, nil

}

func (crm *certRepositoryMock) Save(ctx context.Context, cert certs.Cert) (string, error) {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	crm.counter++
	cert.Serial = strconv.Itoa(crm.counter)
	crm.certs[cert.Serial] = cert
	crm.certsByThing[cert.ThingID] = cert

	return cert.Serial, nil
}

func (crm *certRepositoryMock) Remove(ctx context.Context, serial string) error {
	crm.mu.Lock()
	defer crm.mu.Unlock()
	c, ok := crm.certs[serial]
	if ok {
		delete(crm.certs, serial)
		delete(crm.certsByThing, c.ThingID)
		return nil
	}
	return errRemove
}

func (crm *certRepositoryMock) RetrieveByThing(ctx context.Context, thingID string) (certs.Cert, error) {
	crm.mu.Lock()
	defer crm.mu.Unlock()
	if c, ok := crm.certsByThing[thingID]; ok {
		return c, nil
	}
	return certs.Cert{}, certs.ErrNotFound
}
