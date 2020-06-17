// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"bufio"
	"bytes"
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

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
)

const (
	VAULT_HOST = "http://127.0.0.1:8200/v1/"
	ISSUE_URL  = "pki_int/issue/"
	ROLE       = "example-dot-com"

	VAULT_TOKEN = "s.eN0R5b500gqpP0JEgebhDoth"
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

	errIssueCert = errors.New("failed to issue certificate")
)

var _ Service = (*certsService)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// IssueCert issues certificate for given thing id if access is granted with token
	IssueCert(thingID, token string) (Cert, error)
}

type struct Config {
	logLevel       string
	dbConfig       postgres.Config
	clientTLS      bool
	encKey         []byte
	caCerts        string
	httpPort       string
	serverCert     string
	serverKey      string
	baseURL        string
	thingsPrefix   string
	esThingsURL    string
	esThingsPass   string
	esThingsDB     string
	esURL          string
	esPass         string
	esDB           string
	esConsumerName string
	jaegerURL      string
	authnURL       string
	certsURL 		string
	authnTimeout   time.Duration
}

type certsService struct {
	auth  mainflux.AuthNServiceClient
	certs CertsRepository
	sdk   mfsdk.SDK
	conf  Config
}

type certReq struct {
	CommonName string `json:"common_name"`
	TTL        string `json:"ttl"`
}

type certData struct {
	Certificate    string   `json:"certificate"`
	IssuingCA      string   `json:"issuing_ca"`
	CAChain        []string `json:"ca_chain"`
	PrivateKey     string   `json:"private_key"`
	PrivateKeyType string   `json:"private_key_type"`
	SerialNumber   string   `json:"serial_number"`
}

type certRes struct {
	LeaseID       string   `json:"lease_id"`
	Renewable     string   `json:"renewable"`
	LeaseDuration string   `json:"lease_duration"`
	Warnings      string   `json:"warnings"`
	Auth          string   `json:"auth"`
	CertData      certData `json:"data"`
	Errors        []string `json:"errors"`
}

// New returns new Certs service.
func New(auth mainflux.AuthNServiceClient, certs CertsRepository, sdk mfsdk.SDK, config Config) Service {
	return &certsService{
		certs: certs,
		sdk:   sdk,
		auth:  auth,
		config: config,
	}
}

func (cs *certsService) IssueCert(thingID string, daysValid string, rsaBits int, token string) (Cert, error) {

	// Get the SystemCertPool, continue with an empty pool on error
	thing, err := cs.sdk.Thing(thingID, token)
	if err != nil {
		return Cert{}, errors.Wrap(errIssueCert, err)
	}


	// If certsURL == "" we don't use 3rd party PKI service.
	if cs.config.certsURL == "" {
		c.ClientCert, c.ClientKey, err = cs.certs(th.Key, daysValid, rsaBits)
		if err != nil {
			return Cert{}, errors.Wrap(ErrCertsCreation, err)
		}
		return c, err
	}


	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedLoadingTrustedCA, err)
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	// Trust the augmented cert pool in our client
	config := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            rootCAs,
	}
	tr := &http.Transport{TLSClientConfig: config}
	client := &http.Client{Transport: tr}
	url := cs.getIssueUrl()

	r := certReq{
		CommonName: thing.Key,
		TTL:        "24h",
	}
	d, err := json.Marshal(r)
	if err != nil {
		return Cert{}, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(d))
	if err != nil {
		return Cert{}, err
	}

	req.Header.Add("X-Vault-Token", VAULT_TOKEN)
	resp, err := client.Do(req)
	if err != nil {
		return Cert{}, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return Cert{}, err
		}
		c := certRes{}
		if err := json.Unmarshal(body, &c); err != nil {
			return Cert{}, err
		}
		return Cert{}, errors.Wrap(errIssueCert, errors.New(c.Errors[0]))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Cert{}, err
	}
	defer resp.Body.Close()
	c := Cert{}
	if err := json.Unmarshal(body, &c); err != nil {
		return Cert{}, err
	}

	return c, nil
}

func (cs *certsService) getIssueUrl() string {
	url := VAULT_HOST + ISSUE_URL + ROLE
	return url
}

func (cs *certsService) Cert(thingID, daysValid string, rsaBits int, token string) (Cert, error) {
	var c Cert

	// Check access rights
	th, err := sdk.Thing(thingID, token)
	if err != nil {
		return Cert{}, err
	}
	r := certReq{
		ThingID:  th.ID,
		ThingKey: th.Key,
	}
	d, err := json.Marshal(r)
	if err != nil {
		return Cert{}, err
	}

	// // If certsURL == "" we don't use 3rd party PKI service.
	// if sdk.certsURL == "" {
	// 	c.ClientCert, c.ClientKey, err = sdk.certs(th.Key, daysValid, rsaBits)
	// 	if err != nil {
	// 		return Cert{}, errors.Wrap(ErrCertsCreation, err)
	// 	}
	// 	return c, err
	// }

	res, err := request(http.MethodPost, token, sdk.certsURL, d)
	if err != nil {
		return Cert{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return Cert{}, ErrCerts
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Cert{}, err
	}
	if err := json.Unmarshal(body, &c); err != nil {
		return Cert{}, err
	}
	return c, nil
}

func (cs *certsService) certs(thingKey, daysValid string, rsaBits int) (string, string, error) {
	if sdk.config.certsCA == nil {
		return "", "", errors.Wrap(errFailedCertCreation, errMissingCACertificate)
	}
	if rsaBits == 0 {
		return "", "", errors.Wrap(errFailedCertCreation, ErrRsaBitsValueWrong)
	}
	var priv interface{}
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)

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
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, cs.certsCA, pubKey, cs.certsCert.PrivateKey)
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

type certReq struct {
	ThingID  string `json:"thing_id,omitempty"`
	ThingKey string `json:"thing_key,omitempty"`
}
