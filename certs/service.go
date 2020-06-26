// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mitchellh/mapstructure"
)

var (
	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrMalformedEntity indicates malformed entity specification.
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrConflict indicates that entity with the same ID or external ID already exists.
	ErrConflict = errors.New("entity already exists")

	// ErrThings indicates failure to communicate with Mainflux Things service.
	// It can be due to networking error or invalid/unauthorized request.
	ErrThings = errors.New("failed to receive response from Things service")

	// ErrExternalKeyNotFound indicates a non-existent certs configuration for given external key
	ErrExternalKeyNotFound = errors.New("failed to get certs configuration for given external key")

	// ErrFailedLoadingTrustedCA
	ErrFailedLoadingTrustedCA = errors.New("failed to load trusted certificates")

	// ErrCertsCreation
	ErrCertsCreation = errors.New("failed to create client certificates")

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.New("failed to remove certificate")

	// ErrCACertificateDoesntExist indicates missing CA certificate required for
	// creating mTLS client certificates
	ErrCACertificateDoesntExist = errors.New("CA certificate doesnt exist")

	// ErrCAKeyDoesntExist indicates missing CA private key
	ErrCAKeyDoesntExist = errors.New("CA certificate key doesnt exist")

	// ErrFailedCertCreation failed to create certificate for thing
	ErrFailedCertCreation = errors.New("failed to create client certificate")

	// ErrFailedDateSetting
	ErrFailedDateSetting = errors.New("failed to set date for certificate")

	// ErrRsaBitsValueWrong
	ErrRsaBitsValueWrong = errors.New("missing RSA bits for certificate creation")

	// ErrMissingCACertificate
	ErrMissingCACertificate = errors.New("missing CA certificate for certificate signing")

	// ErrFailedSerialGeneration
	ErrFailedSerialGeneration = errors.New("failed to generate certificate serial")

	// ErrFailedPemKeyWrite
	ErrFailedPemKeyWrite = errors.New("failed to write PEM key")

	// ErrFailedPemDataWrite
	ErrFailedPemDataWrite = errors.New("failed to write pem data for certificate")

	// ErrPrivateKeyUnsupportedType
	ErrPrivateKeyUnsupportedType = errors.New("private key type is unsupported")

	// ErrPrivateKeyEmpty
	ErrPrivateKeyEmpty = errors.New("private key is empty")
)

var _ Service = (*certsService)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// IssueCert issues certificate for given thing id if access is granted with token
	IssueCert(thingID, daysValid string, keyBits int, keyType string, token string) (Cert, error)
}

type Config struct {
	LogLevel       string
	ClientTLS      bool
	CaCerts        string
	HttpPort       string
	ServerCert     string
	ServerKey      string
	BaseURL        string
	ThingsPrefix   string
	EsThingsURL    string
	EsThingsPass   string
	EsThingsDB     string
	EsURL          string
	EsPass         string
	EsDB           string
	EsConsumerName string
	JaegerURL      string
	AuthnURL       string
	AuthnTimeout   time.Duration
	SignTLSCert    tls.Certificate
	SignX509Cert   *x509.Certificate
	SignRSABits    int
	SignHoursValid string

	PkiHost        string
	PkiIssueURL    string
	PkiAccessToken string
	PkiRole        string
}

type certsService struct {
	auth      mainflux.AuthNServiceClient
	certsRepo CertsRepository
	sdk       mfsdk.SDK
	conf      Config
	pkiClient *api.Client
}

type Cert struct {
	ThingID        string        `mapstructure:"-"`
	ClientCert     string        `mapstructure:"certificate"`
	IssuingCA      string        `mapstructure:"issuing_ca"`
	CAChain        []string      `mapstructure:"ca_chain"`
	ClientKey      string        `mapstructure:"private_key"`
	PrivateKeyType string        `mapstructure:"private_key_type"`
	Serial         string        `mapstructure:"serial_number"`
	Expire         time.Duration `mapstructure:"expiration"`
}

type pkiResponse struct {
	ClientCert     string        `mapstructure:"certificate"`
	IssuingCA      string        `mapstructure:"issuing_ca"`
	CAChain        []string      `mapstructure:"ca_chain"`
	ClientKey      string        `mapstructure:"private_key"`
	PrivateKeyType string        `mapstructure:"private_key_type"`
	Serial         string        `mapstructure:"serial_number"`
	Expire         time.Duration `mapstructure:"expiration"`
}

type certReq struct {
	CommonName string `json:"common_name"`
	TTL        string `json:"ttl"`
	KeyBits    int    `json:"key_bits"`
	KeyType    string `json:"key_type"`
}

type vaultRes struct {
	Data map[string]interface{} `json:"data" mapstructure:"data"`
}

// New returns new Certs service.
func New(auth mainflux.AuthNServiceClient, certs CertsRepository, sdk mfsdk.SDK, config Config, c *api.Client) Service {
	return &certsService{
		certsRepo: certs,
		sdk:       sdk,
		auth:      auth,
		conf:      config,
		pkiClient: c,
	}
}

func (cs *certsService) IssueCert(thingID string, daysValid string, keyBits int, keyType string, token string) (Cert, error) {
	thing, err := cs.sdk.Thing(thingID, token)
	if err != nil {
		return Cert{}, errors.Wrap(ErrCertsCreation, err)
	}
	var c Cert

	// If pkiClient == nil we don't use 3rd party PKI service.
	if cs.pkiClient == nil {
		c.ClientCert, c.ClientKey, err = cs.certs(thing.Key, daysValid, keyBits)
		if err != nil {
			return Cert{}, errors.Wrap(ErrCertsCreation, err)
		}
		return c, err
	}

	cReq := certReq{
		CommonName: thing.Key,
		TTL:        daysValid,
		KeyBits:    keyBits,
		KeyType:    keyType,
	}

	r := cs.pkiClient.NewRequest("POST", cs.getIssueUrl())
	if err := r.SetJSONBody(cReq); err != nil {
		return Cert{}, err
	}

	resp, err := cs.pkiClient.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return Cert{}, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		_, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return Cert{}, err
		}
		return Cert{}, errors.Wrap(ErrCertsCreation, err)
	}

	s, _ := api.ParseSecret(resp.Body)
	pr := pkiResponse{}

	if err := mapstructure.Decode(s.Data, &pr); err != nil {
		return Cert{}, err
	}

	c = Cert{
		ClientCert:     pr.ClientCert,
		ClientKey:      pr.ClientKey,
		PrivateKeyType: pr.PrivateKeyType,
		IssuingCA:      pr.IssuingCA,
		Serial:         pr.Serial,
		CAChain:        pr.CAChain,
		ThingID:        thing.ID,
	}
	_, err = cs.certsRepo.Save(context.Background(), c)
	return c, err
}

func (cs *certsService) RetrieveAll ()
func (cs *certsService) getIssueUrl() string {
	url := cs.conf.PkiIssueURL + cs.conf.PkiRole
	return url
}

func (cs *certsService) certs(thingKey, daysValid string, keyBits int) (string, string, error) {
	if cs.conf.SignX509Cert == nil {
		return "", "", errors.Wrap(ErrFailedCertCreation, ErrMissingCACertificate)
	}
	if keyBits == 0 {
		return "", "", errors.Wrap(ErrFailedCertCreation, ErrRsaBitsValueWrong)
	}
	var priv interface{}
	// p224 := elliptic.P224()
	// priv, err := elliptic.GenerateKey(p224, rand.Reader)
	priv, err := rsa.GenerateKey(rand.Reader, keyBits)

	if daysValid == "" {
		daysValid = cs.conf.SignHoursValid
	}

	notBefore := time.Now()
	validFor, err := time.ParseDuration(daysValid)
	if err != nil {
		return "", "", errors.Wrap(ErrFailedDateSetting, err)
	}
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", errors.Wrap(ErrFailedSerialGeneration, err)
	}

	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"Mainflux"},
			CommonName:         thingKey,
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
		return "", "", errors.Wrap(ErrFailedCertCreation, err)
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, cs.conf.SignX509Cert, pubKey, cs.conf.SignTLSCert.PrivateKey)
	if err != nil {
		return "", "", errors.Wrap(ErrFailedCertCreation, err)
	}

	var bw, keyOut bytes.Buffer
	buffWriter := bufio.NewWriter(&bw)
	buffKeyOut := bufio.NewWriter(&keyOut)

	if err := pem.Encode(buffWriter, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", errors.Wrap(ErrFailedPemDataWrite, err)
	}
	buffWriter.Flush()
	cert := bw.String()

	block, err := pemBlockForKey(priv)
	if err != nil {
		return "", "", errors.Wrap(ErrFailedPemKeyWrite, err)
	}
	if err := pem.Encode(buffKeyOut, block); err != nil {
		return "", "", errors.Wrap(ErrFailedPemKeyWrite, err)
	}
	buffKeyOut.Flush()
	key := keyOut.String()

	return cert, key, nil
}

func publicKey(priv interface{}) (interface{}, error) {
	if priv == nil {
		return nil, ErrPrivateKeyEmpty
	}
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey, nil
	case *ecdsa.PrivateKey:
		return &k.PublicKey, nil
	default:
		return nil, ErrPrivateKeyUnsupportedType
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

func request(method, jwt, url string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", jwt)
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
