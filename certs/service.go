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

	// ErrConflict indicates that entity with the same ID or external ID already exists.
	ErrConflict = errors.New("entity already exists")

	// ErrThings indicates failure to communicate with Mainflux Things service.
	// It can be due to networking error or invalid/unauthorized request.
	ErrThings = errors.New("failed to receive response from Things service")

	// ErrExternalKeyNotFound indicates a non-existent certs configuration for given external key
	ErrExternalKeyNotFound = errors.New("failed to get certs configuration for given external key")

	// ErrFailedLoadingTrustedCA indicates problem with trusted certificates
	ErrFailedLoadingTrustedCA = errors.New("failed to load trusted certificates")

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.New("failed to remove certificate")

	// ErrCACertificateDoesntExist indicates missing CA certificate required for
	// creating mTLS client certificates
	ErrCACertificateDoesntExist = errors.New("CA certificate doesnt exist")

	// ErrCAKeyDoesntExist indicates missing CA private key
	ErrCAKeyDoesntExist = errors.New("CA certificate key doesnt exist")

	// ErrFailedCertCreation indicates problem in certificate creation
	ErrFailedCertCreation = errors.New("failed to create client certificate")

	// ErrFailedDateSetting failed to set date for certificate
	ErrFailedDateSetting = errors.New("failed to set date for certificate")

	// ErrRsaBitsValueWrong indicates missing RSA bits for certificate creation
	ErrRsaBitsValueWrong = errors.New("missing RSA bits for certificate creation")

	// ErrMissingCACertificate indicates missing CA certificate for certificate signing
	ErrMissingCACertificate = errors.New("missing CA certificate for certificate signing")

	// ErrFailedSerialGeneration failed to generate certificate serial
	ErrFailedSerialGeneration = errors.New("failed to generate certificate serial")

	// ErrFailedPemKeyWrite indicates problem with writing PEM key
	ErrFailedPemKeyWrite = errors.New("failed to write PEM key")

	// ErrFailedPemDataWrite failed to write pem data for certificate
	ErrFailedPemDataWrite = errors.New("failed to write pem data for certificate")

	// ErrPrivateKeyUnsupportedType indicates problem with unsupported  private key type
	ErrPrivateKeyUnsupportedType = errors.New("private key type is unsupported")

	// ErrPrivateKeyEmpty indicates private key empty
	ErrPrivateKeyEmpty = errors.New("private key is empty")

	// ErrMissingCertSerial indicates problem with missing certificate serial
	ErrMissingCertSerial = errors.New("missing cert serial")

	// ErrFailedToRemoveCertFromDB indicates problem in removing cert from db
	ErrFailedToRemoveCertFromDB = errors.New("failed to remove cert serial from db")

	// ErrFailedToParseCertificate indicates problem parsing certificate
	ErrFailedToParseCertificate = errors.New("failed to parse x509 certificate")
)

var _ Service = (*certsService)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// IssueCert issues certificate for given thing id if access is granted with token
	IssueCert(ctx context.Context, token, thingID, daysValid string, keyBits int, keyType string) (Cert, error)

	// ListCertificates lists all certificates issued for given thing
	ListCertificates(ctx context.Context, token, thingID string, offset, limit uint64) (CertsPage, error)

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
	certsRepo CertsRepository
	sdk       mfsdk.SDK
	conf      Config
	PKIClient *api.Client
}

type Cert struct {
	ThingID        string    `json:"thing_id" mapstructure:"-"`
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
		PKIClient: c,
	}
}

func (cs *certsService) IssueCert(ctx context.Context, token, thingID string, daysValid string, keyBits int, keyType string) (Cert, error) {
	thing, err := cs.sdk.Thing(thingID, token)
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedCertCreation, err)
	}
	var c Cert

	// If PKIClient == nil we don't use 3rd party PKI service.
	if cs.conf.PKIHost == "" {
		c.ClientCert, c.ClientKey, err = cs.certs(thing.Key, daysValid, keyBits)
		if err != nil {
			return Cert{}, errors.Wrap(ErrFailedCertCreation, err)
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
		return Cert{}, errors.Wrap(ErrFailedCertCreation, err)
	}

	s, _ := api.ParseSecret(resp.Body)
	cert := Cert{}

	mapstructure.Decode(s.Data, &cert)

	// Expire time calc must be revised
	// value doesnt look correct
	exp, err := s.Data["expiration"].(json.Number).Float64()
	expTime := time.Unix(0, int64(exp)*int64(time.Millisecond))
	cert.Expire = expTime

	cert.ThingID = thing.ID

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
		return Revoke{}, errors.Wrap(ErrFailedCertCreation, err)
	}

	s, err := api.ParseSecret(resp.Body)
	if err != nil {
		return Revoke{}, err
	}

	rev, err := s.Data["revocation_time"].(json.Number).Float64()
	revTime := time.Unix(0, int64(rev)*int64(time.Millisecond))
	revoke := Revoke{
		RevocationTime: revTime,
	}

	c := Cert{
		Serial: certSerial,
	}

	if err = cs.certsRepo.Remove(context.Background(), c); err != nil {
		return Revoke{}, errors.Wrap(ErrFailedToRemoveCertFromDB, err)
	}
	return revoke, nil

}

func (cs *certsService) ListCertificates(ctx context.Context, token, thingID string, offset, limit uint64) (CertsPage, error) {
	_, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return CertsPage{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return cs.certsRepo.RetrieveAll(ctx, thingID, offset, limit)
}

func (cs *certsService) getIssueURL() string {
	return cs.conf.PKIHost + "/" + apiVer + "/" + cs.conf.PKIPath + "/" + issue + "/" + cs.conf.PKIRole
}

func (cs *certsService) getRevokeURL() string {
	return cs.conf.PKIHost + "/" + apiVer + "/" + cs.conf.PKIPath + "/" + revoke
}

func (cs *certsService) certs(thingKey, daysValid string, keyBits int) (string, string, error) {
	if cs.conf.SignX509Cert == nil {
		return "", "", errors.Wrap(ErrFailedCertCreation, ErrMissingCACertificate)
	}
	if keyBits == 0 {
		return "", "", errors.Wrap(ErrFailedCertCreation, ErrRsaBitsValueWrong)
	}
	var priv interface{}
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
