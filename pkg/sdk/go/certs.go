package sdk

import (
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	ErrCertsCreation     = errors.New("failed to create certificate")
	ErrRsaBitsValueWrong = errors.New("value for RSA bits must be > 0")

	errFailedCertCreation     = errors.New("failed creating certificate")
	errFailedDateSetting      = errors.New("failed setting date")
	errFailedPemDataWrite     = errors.New("failed writing pem data")
	errFailedPemKeyWrite      = errors.New("failed writing pem key data")
	errFailedSerialGeneration = errors.New("failed generating certificates serial")
)

// Cert represents certs data.
type Cert struct {
	CACert     string `json:"ca_cert,omitempty"`
	ClientKey  string `json:"client_key,omitempty"`
	ClientCert string `json:"client_cert,omitempty"`
}
