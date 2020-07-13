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
	"encoding/json"
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

const (
	issue  = "issue"
	revoke = "revoke"
	apiVer = "v1"
)

var (
	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrMalformedEntity indicates malformed entity specification.
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrMissingCertSerial indicates problem with missing certificate serial
	ErrMissingCertSerial = errors.New("missing cert serial")

	errFailedCertCreation        = errors.New("failed to create client certificate")
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
	errFailedCertDecoding        = errors.New("failed to decode response from PKI service")
)

var _ Service = (*certsService)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// IssueCert issues certificate for given thing id if access is granted with token
	IssueCert(ctx context.Context, token, thingID, daysValid string, keyBits int, keyType string) (Cert, error)

	// ListCerts lists all certificates issued for given owner
	ListCerts(ctx context.Context, token string, offset, limit uint64) (Page, error)

	// RevokeCert
	RevokeCert(ctx context.Context, token, thingID, certID string) (Revoke, error)
}

type Config struct {
	LogLevel       string
	ClientTLS      bool
	CaCerts        string
	HTTPPort       string
	ServerCert     string
	ServerKey      string
	BaseURL        string
	ThingsPrefix   string
	JaegerURL      string
	AuthnURL       string
	AuthnTimeout   time.Duration
	SignTLSCert    tls.Certificate
	SignX509Cert   *x509.Certificate
	SignRSABits    int
	SignHoursValid string
	PKIHost        string
	PKIPath        string
	PKIRole        string
	PKIToken       string
}

type certsService struct {
	auth      mainflux.AuthNServiceClient
	certsRepo Repository
	sdk       mfsdk.SDK
	conf      Config
	PKIClient *api.Client
}

type Cert struct {
	ThingID        string    `json:"thing_id" mapstructure:"-"`
	OwnerID        string    `json:"-" mapstructure:"-"`
	ClientCert     string    `json:"client_cert" mapstructure:"certificate"`
	IssuingCA      string    `json:"issuing_ca" mapstructure:"issuing_ca"`
	CAChain        []string  `json:"ca_chain" mapstructure:"ca_chain"`
	ClientKey      string    `json:"client_key" mapstructure:"private_key"`
	PrivateKeyType string    `json:"private_key_type" mapstructure:"private_key_type"`
	Serial         string    `json:"serial" mapstructure:"serial_number"`
	Expire         time.Time `json:"expire" mapstructure:"-"`
}

type Revoke struct {
	RevocationTime time.Time `mapstructure:"revocation_time"`
}

type certReq struct {
	CommonName string `json:"common_name"`
	TTL        string `json:"ttl"`
	KeyBits    int    `json:"key_bits"`
	KeyType    string `json:"key_type"`
}

type certRevokeReq struct {
	SerialNumber string `json:"serial_number"`
}

// New returns new Certs service.
func New(auth mainflux.AuthNServiceClient, certs Repository, sdk mfsdk.SDK, config Config, c *api.Client) Service {
	return &certsService{
		certsRepo: certs,
		sdk:       sdk,
		auth:      auth,
		conf:      config,
		PKIClient: c,
	}
}

func (cs *certsService) IssueCert(ctx context.Context, token, thingID string, daysValid string, keyBits int, keyType string) (Cert, error) {
	owner, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Cert{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	thing, err := cs.sdk.Thing(thingID, token)
	if err != nil {
		return Cert{}, errors.Wrap(errFailedCertCreation, err)
	}
	var c Cert

	// If PKIClient == nil we don't use 3rd party PKI service.
	if cs.conf.PKIHost == "" {
		c.ClientCert, c.ClientKey, err = cs.certs(thing.Key, daysValid, keyBits)
		if err != nil {
			return Cert{}, errors.Wrap(errFailedCertCreation, err)
		}
		return c, err
	}

	cReq := certReq{
		CommonName: thing.Key,
		TTL:        daysValid,
		KeyBits:    keyBits,
		KeyType:    keyType,
	}

	r := cs.PKIClient.NewRequest("POST", cs.getIssueURL())
	if err := r.SetJSONBody(cReq); err != nil {
		return Cert{}, err
	}

	resp, err := cs.PKIClient.RawRequest(r)
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
		return Cert{}, errors.Wrap(errFailedCertCreation, err)
	}

	s, _ := api.ParseSecret(resp.Body)
	cert := Cert{}

	if err = mapstructure.Decode(s.Data, &cert); err != nil {
		return Cert{}, errors.Wrap(errFailedCertDecoding, err)
	}

	// Expire time calc must be revised
	// value doesnt look correct
	exp, err := s.Data["expiration"].(json.Number).Float64()
	if err != nil {
		return cert, err
	}
	expTime := time.Unix(0, int64(exp)*int64(time.Millisecond))
	cert.Expire = expTime
	cert.ThingID = thing.ID
	cert.OwnerID = owner.GetValue()

	_, err = cs.certsRepo.Save(context.Background(), cert)
	return cert, err
}

func (cs *certsService) RevokeCert(ctx context.Context, token, thingID, certSerial string) (Revoke, error) {
	_, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Revoke{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	cReq := certRevokeReq{
		SerialNumber: certSerial,
	}

	r := cs.PKIClient.NewRequest("POST", cs.getRevokeURL())
	if err := r.SetJSONBody(cReq); err != nil {
		return Revoke{}, err
	}

	resp, err := cs.PKIClient.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return Revoke{}, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		_, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return Revoke{}, err
		}
		return Revoke{}, errors.Wrap(errFailedCertCreation, err)
	}

	s, err := api.ParseSecret(resp.Body)
	if err != nil {
		return Revoke{}, err
	}

	rev, err := s.Data["revocation_time"].(json.Number).Float64()
	if err != nil {
		return Revoke{}, err
	}
	revTime := time.Unix(0, int64(rev)*int64(time.Millisecond))
	revoke := Revoke{
		RevocationTime: revTime,
	}

	c := Cert{
		Serial: certSerial,
	}

	if err = cs.certsRepo.Remove(context.Background(), c); err != nil {
		return Revoke{}, errors.Wrap(errFailedToRemoveCertFromDB, err)
	}
	return revoke, nil

}

func (cs *certsService) ListCerts(ctx context.Context, token string, offset, limit uint64) (Page, error) {
	u, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Page{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return cs.certsRepo.RetrieveAll(ctx, u.GetValue(), offset, limit)
}

func (cs *certsService) getIssueURL() string {
	return "/" + apiVer + "/" + cs.conf.PKIPath + "/" + issue + "/" + cs.conf.PKIRole
}

func (cs *certsService) getRevokeURL() string {
	return "/" + apiVer + "/" + cs.conf.PKIPath + "/" + revoke
}

func (cs *certsService) certs(thingKey, daysValid string, keyBits int) (string, string, error) {
	if cs.conf.SignX509Cert == nil {
		return "", "", errors.Wrap(errFailedCertCreation, errMissingCACertificate)
	}
	if keyBits == 0 {
		return "", "", errors.Wrap(errFailedCertCreation, errKeyBitsValueWrong)
	}
	var priv interface{}
	priv, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return "", "", errors.Wrap(errFailedKeyCreation, err)
	}

	if daysValid == "" {
		daysValid = cs.conf.SignHoursValid
	}

	notBefore := time.Now()
	validFor, err := time.ParseDuration(daysValid)
	if err != nil {
		return "", "", errors.Wrap(errFailedDateSetting, err)
	}
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", errors.Wrap(errFailedSerialGeneration, err)
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
		return "", "", errors.Wrap(errFailedCertCreation, err)
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, cs.conf.SignX509Cert, pubKey, cs.conf.SignTLSCert.PrivateKey)
	if err != nil {
		return "", "", errors.Wrap(errFailedCertCreation, err)
	}

	var bw, keyOut bytes.Buffer
	buffWriter := bufio.NewWriter(&bw)
	buffKeyOut := bufio.NewWriter(&keyOut)

	if err := pem.Encode(buffWriter, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", errors.Wrap(errFailedPemDataWrite, err)
	}
	buffWriter.Flush()
	cert := bw.String()

	block, err := pemBlockForKey(priv)
	if err != nil {
		return "", "", errors.Wrap(errFailedPemKeyWrite, err)
	}
	if err := pem.Encode(buffKeyOut, block); err != nil {
		return "", "", errors.Wrap(errFailedPemKeyWrite, err)
	}
	buffKeyOut.Flush()
	key := keyOut.String()

	return cert, key, nil
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
