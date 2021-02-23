// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package pki wraps native cert agent
package pki

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
)

var _ Agent = (*agent)(nil)

var (
	//Indicate that method called is not implemented
	ErrNotImplemented = errors.New("method not implemented for certs")

	errFailedKeyCreation         = errors.New("failed to create client private key")
	errFailedDateSetting         = errors.New("failed to set date for certificate")
	errKeyBitsValueWrong         = errors.New("missing RSA bits for certificate creation")
	errMissingCACertificate      = errors.New("missing CA certificate for certificate signing")
	errFailedSerialGeneration    = errors.New("failed to generate certificate serial")
	errFailedPemKeyWrite         = errors.New("failed to write PEM key")
	errFailedPemDataWrite        = errors.New("failed to write pem data for certificate")
	errPrivateKeyUnsupportedType = errors.New("private key type is unsupported")
	errPrivateKeyEmpty           = errors.New("private key is empty")
	errFailedToRemoveCertFromDB  = errors.New("failed to remove cert serial from db")
	errFailedCertCreation        = errors.New("failed to create client certificate")
	errFailedCertRevocation      = errors.New("failed to revoke certificate")
)

type agent struct {
	AuthTimeout    time.Duration
	SignTLSCert    tls.Certificate
	SignX509Cert   *x509.Certificate
	SignRSABits    int
	SignHoursValid string
}
type Revoke struct {
	RevocationTime time.Time `mapstructure:"revocation_time"`
}

type Cert struct {
	ClientCert     string    `json:"client_cert" mapstructure:"certificate"`
	IssuingCA      string    `json:"issuing_ca" mapstructure:"issuing_ca"`
	CAChain        []string  `json:"ca_chain" mapstructure:"ca_chain"`
	ClientKey      string    `json:"client_key" mapstructure:"private_key"`
	PrivateKeyType string    `json:"private_key_type" mapstructure:"private_key_type"`
	Serial         string    `json:"serial" mapstructure:"serial_number"`
	Expire         time.Time `json:"expire" mapstructure:"-"`
}

type Agent interface {
	// IssueCert issues certificate on PKI
	IssueCert(cn string, ttl, keyType string, keyBits int) (Cert, error)
	// Revoke revokes certificate from PKI
	Revoke(serial string) (Revoke, error)
}

func NewAgent() Agent {
	return &agent{}
}

func (a *agent) IssueCert(cn string, ttl, keyType string, keyBits int) (Cert, error) {
	return a.certs(cn, ttl, keyBits)
}

func (a *agent) certs(cn, daysValid string, keyBits int) (Cert, error) {
	if a.SignX509Cert == nil {
		return Cert{}, errors.Wrap(errFailedCertCreation, errMissingCACertificate)
	}
	if keyBits == 0 {
		return Cert{}, errors.Wrap(errFailedCertCreation, errKeyBitsValueWrong)
	}
	var priv interface{}
	priv, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return Cert{}, errors.Wrap(errFailedKeyCreation, err)
	}

	if daysValid == "" {
		daysValid = a.SignHoursValid
	}

	notBefore := time.Now()
	validFor, err := time.ParseDuration(daysValid)
	if err != nil {
		return Cert{}, errors.Wrap(errFailedDateSetting, err)
	}
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return Cert{}, errors.Wrap(errFailedSerialGeneration, err)
	}

	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"Mainflux"},
			CommonName:         cn,
			OrganizationalUnit: []string{"mainflux"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
	}

	pubKey, err := publicKey(priv)
	if err != nil {
		return Cert{}, errors.Wrap(errFailedCertCreation, err)
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, a.SignX509Cert, pubKey, a.SignTLSCert.PrivateKey)
	if err != nil {
		return Cert{}, errors.Wrap(errFailedCertCreation, err)
	}

	x509cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return Cert{}, errors.Wrap(errFailedCertCreation, err)
	}

	var bw, keyOut bytes.Buffer
	buffWriter := bufio.NewWriter(&bw)
	buffKeyOut := bufio.NewWriter(&keyOut)

	if err := pem.Encode(buffWriter, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return Cert{}, errors.Wrap(errFailedPemDataWrite, err)
	}
	buffWriter.Flush()
	cert := bw.String()

	block, err := pemBlockForKey(priv)
	if err != nil {
		return Cert{}, errors.Wrap(errFailedPemKeyWrite, err)
	}
	if err := pem.Encode(buffKeyOut, block); err != nil {
		return Cert{}, errors.Wrap(errFailedPemKeyWrite, err)
	}
	buffKeyOut.Flush()
	key := keyOut.String()

	return Cert{
		ClientCert: cert,
		ClientKey:  key,
		Serial:     x509cert.Subject.SerialNumber,
		Expire:     x509cert.NotAfter,
		IssuingCA:  x509cert.Issuer.String(),
	}, nil
}

func publicKey(priv interface{}) (interface{}, error) {
	if priv == nil {
		return nil, errPrivateKeyEmpty
	}
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey, nil
	case *ecdsa.PrivateKey:
		return &k.PublicKey, nil
	default:
		return nil, errPrivateKeyUnsupportedType
	}
}

func pemBlockForKey(priv interface{}) (*pem.Block, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, nil
	}
}

func (a *agent) Revoke(serial string) (Revoke, error) {

	return Revoke{}, ErrNotImplemented

}
