package mocks

import (
	"context"
	"sync"

	"github.com/mainflux/mainflux/certs"
)

var _ certs.Repository = (*certsRepoMock)(nil)

type certsRepoMock struct {
	mu             sync.Mutex
	counter        uint64
	certs          map[string]certs.Cert
	certsByThingID map[string]certs.Cert
}

// NewCertsRepository creates in-memory certs repository.
func NewCertsRepository() certs.Repository {
	return &certsRepoMock{
		certs:          make(map[string]certs.Cert),
		certsByThingID: make(map[string]certs.Cert),
	}
}

func (c *certsRepoMock) Save(ctx context.Context, cert certs.Cert) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.certs[cert.Serial] = cert
	c.certsByThingID[cert.ThingID] = cert
	c.counter++
	return cert.Serial, nil
}

func (c *certsRepoMock) RetrieveAll(ctx context.Context, ownerID, thingID string, offset, limit uint64) (certs.Page, error) {
	defer c.mu.Unlock()
	if limit <= 0 {
		return certs.Page{}, nil
	}

	var crts []certs.Cert
	for _, v := range c.certs {
		crts = append(crts, v)
	}

	page := certs.Page{
		Certs:  crts,
		Total:  c.counter,
		Offset: offset,
		Limit:  limit,
	}
	return page, nil

}

func (c *certsRepoMock) Remove(ctx context.Context, thingID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	crt, ok := c.certsByThingID[thingID]
	if !ok {
		return certs.ErrNotFound
	}
	delete(c.certs, crt.Serial)
	delete(c.certsByThingID, thingID)
	return nil
}

func (c *certsRepoMock) RetrieveByThing(ctx context.Context, thingID string) (certs.Cert, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	crt, ok := c.certsByThingID[thingID]
	if !ok {
		return certs.Cert{}, certs.ErrNotFound
	}
	return crt, nil
}
