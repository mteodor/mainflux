package mocks

import "github.com/mainflux/mainflux/certs/pki"

var _ pki.Agent = (*pkiAgentMock)(nil)

type pkiAgentMock struct {
}

func new PkiAgent() pki.Agent {
	return pki.NewAgent()
}
