package mocks

import (
	"context"
	"sync"

	"github.com/mainflux/mainflux/certs"
)

var _ certs.Repository = (*certsRepoMock)(nil)

type certsRepoMock struct {
	mu      sync.Mutex
	counter uint64
	certs   map[string]certs.Cert
}

// NewCertsRepository creates in-memory certs repository.
func NewCertsRepository() certs.Cert {
	return &configRepositoryMock{
		certs: make(map[string]certs.Cert),
	}
}

func (c *certsRepoMock) Save(ctx context.Context, cert certs.Cert) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	certs[cert.Serial] = cert
	return cert.Serial, nil
}

func (c *certsRepoMock) RetrieveAll(ctx context.Context, ownerID, thingID string, offset, limit uint64) (certs.Page, error) {

}

func (c *certsRepoMock) Remove(ctx context.Context, thingID string) error {

}

func (c *certsRepoMock) RetrieveByThing(ctx context.Context, thingID string) (certs.Cert, error) {

}
