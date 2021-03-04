package mocks

import "github.com/mainflux/mainflux/certs/pki"

func NewPkiAgent() pki.Agent {
	return pki.NewAgent()
}
