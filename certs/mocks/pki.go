package mocks

import "github.com/mainflux/mainflux/certs/pki"

var _ pki.Agent = (*pkiAgentMock)(nil)

func NewPkiAgent() pki.Agent {
	return pki.NewAgent()
}
