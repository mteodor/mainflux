package certs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mainflux/mainflux/errors"
	mfSDK "github.com/mainflux/mainflux/sdk/go"
)

// Thing is Mainflux SDK thing.
type Thing mfSDK.Thing

// Channel is Mainflux SDK channel.
type Channel mfSDK.Channel

var (
	// ErrCerts indicates error fetching certificates.
	ErrCerts = errors.New("failed to fetch certs data")

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.New("failed to remove certificate")

	// ErrConflict indicates duplicate unique field.
	ErrConflict = errors.New("duplicate unique field")

	// ErrUnauthorized indicates forbidden access.
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrMalformedEntity indicates malformed request data.
	ErrMalformedEntity = errors.New("malformed data")

	// ErrNotFound indicates that entity doesn't exist.
	ErrNotFound = errors.New("entity not found")
)

// BSConfig represents Config entity to be stored by Bootstrap service.
type BSConfig struct {
	ThingID     string   `json:"thing_id,omitempty"`
	ExternalID  string   `json:"external_id,omitempty"`
	ExternalKey string   `json:"external_key,omitempty"`
	Channels    []string `json:"channels,omitempty"`
	Content     string   `json:"content,omitempty"`
	ClientCert  string   `json:"client_cert,omitempty"`
	ClientKey   string   `json:"client_key,omitempty"`
	CACert      string   `json:"ca_cert,omitempty"`
}

// Cert represents certs data.
type Cert struct {
	CACert     string `json:"ca_cert,omitempty"`
	ClientKey  string `json:"client_key,omitempty"`
	ClientCert string `json:"client_cert,omitempty"`
}

// SDK is wrapper around cert SDK that adds certs management.
type SDK interface {

	// Cert creates cert using external PKI provider and Certs service.
	Cert(thingID, thingKey, token string) (Cert, error)

	// RemoveCert revokes and removes cert from Certs service.
	RemoveCert(key, token string) error
}

type provisionSDK struct {
	certsURL string
}

// New creates new Provision SDK.
func New(certsURL string) SDK {
	return &provisionSDK{
		certsURL: certsURL,
	}
}
func (ps *provisionSDK) Cert(thingID, thingKey, token string) (Cert, error) {
	var c Cert
	r := certReq{
		ThingID:  thingID,
		ThingKey: thingKey,
	}
	d, err := json.Marshal(r)
	if err != nil {
		return Cert{}, err
	}
	res, err := request(http.MethodPost, token, ps.certsURL, d)
	if err != nil {
		return Cert{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return Cert{}, ErrCerts
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		println(err.Error())
		return Cert{}, err
	}
	if err := json.Unmarshal(body, &c); err != nil {
		return Cert{}, err
	}
	return c, nil
}

func (ps *provisionSDK) RemoveCert(id, token string) error {
	res, err := request(http.MethodDelete, token, fmt.Sprintf("%s/%s", ps.certsURL, id), nil)
	if res != nil {
		res.Body.Close()
	}
	if err != nil {
		return err
	}
	switch res.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusForbidden:
		return ErrUnauthorized
	default:
		return ErrCertsRemove
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
	ThingID  string `json:"id,omitempty"`
	ThingKey string `json:"key,omitempty"`
}
