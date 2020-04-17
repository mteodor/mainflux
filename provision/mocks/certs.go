package mocks

//
import (
	"sync"

	"github.com/mainflux/mainflux/provision/certs"
	mfSDK "github.com/mainflux/mainflux/sdk/go"
)

const (
	validEmail   = "test@example.com"
	validPass    = "test"
	invalid      = "invalid"
	validToken   = "valid_token"
	invalidToken = "invalid_token"
)

var thingIDs = []string{"ids"}

// SDK is fake sdk for mocking
type mockCerts struct {
	things      map[string]mfSDK.Thing
	channels    map[string]mfSDK.Channel
	connections map[string][]string
	configs     map[string]mfSDK.BoostrapConfig
	mu          sync.Mutex
}

// NewSDK returns new mock SDK for testing purposes.
func NewCertsSDK() certs.SDK {
	sdk := &mockCerts{}
	th := mfSDK.Thing{ID: "predefined", Name: "ID"}
	sdk.things = map[string]mfSDK.Thing{"predefined": th}
	sdk.mu = sync.Mutex{}
	return sdk
}

func (s *mockCerts) Cert(thingID, thingKey string, token string) (certs.Cert, error) {
	if thingID == invalid || thingKey == invalid {
		return certs.Cert{}, certs.ErrCerts
	}
	return certs.Cert{}, nil
}

func (s *mockCerts) RemoveCert(key string, token string) error {
	if key == invalid {
		return certs.ErrCertsRemove
	}
	return nil
}
