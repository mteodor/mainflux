// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
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

type certsService struct {
	auth  mainflux.AuthNServiceClient
	certs CertsRepository
	sdk   mfsdk.SDK
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
func New(auth mainflux.AuthNServiceClient, certs CertsRepository, sdk mfsdk.SDK) Service {
	return &certsService{
		certs: certs,
		sdk:   sdk,
		auth:  auth,
	}
}

func (cs *certsService) IssueCert(thingID string, token string) (Cert, error) {

	// Get the SystemCertPool, continue with an empty pool on error
	thing, err := cs.sdk.Thing(thingID, token)
	if err != nil {
		return Cert{}, errors.Wrap(errIssueCert, err)
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
