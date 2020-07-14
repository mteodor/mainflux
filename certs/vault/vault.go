package vault

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mitchellh/mapstructure"
)

const (
	issue  = "issue"
	revoke = "revoke"
	apiVer = "v1"
)

var (
	errFailedVaultCertIssue = errors.New("failed to issue vault certificate")
	errFailedCertDecoding   = errors.New("failed to decode response from vault service")
)

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

type pki struct {
	issueURL string
	token    string
	path     string
	role     string
	host     string
	client   *api.Client
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

func NewVaultClient(token, host, path, role string) (*pki, error) {
	conf := &api.Config{
		Address: host,
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	p := pki{
		token:  token,
		host:   host,
		role:   role,
		path:   path,
		client: client,
	}
	return &p, nil
}

func (p *pki) IssueCert(cn string, ttl, keyType string, keyBits int) (Cert, error) {
	cReq := certReq{
		CommonName: cn,
		TTL:        ttl,
		KeyBits:    keyBits,
		KeyType:    keyType,
	}

	r := p.client.NewRequest("POST", p.getIssueURL())
	if err := r.SetJSONBody(cReq); err != nil {
		return Cert{}, err
	}

	resp, err := p.client.RawRequest(r)
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
		return Cert{}, errors.Wrap(errFailedVaultCertIssue, err)
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
	return cert, nil

}

func (p *pki) Revoke(serial string) (Revoke, error) {
	cReq := certRevokeReq{
		SerialNumber: serial,
	}

	r := p.client.NewRequest("POST", p.getRevokeURL())
	if err := r.SetJSONBody(cReq); err != nil {
		return Revoke{}, err
	}

	resp, err := p.client.RawRequest(r)
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
		return Revoke{}, errors.Wrap(errFailedVaultCertIssue, err)
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
	return Revoke{
		RevocationTime: revTime,
	}, nil

}

func (p *pki) getIssueURL() string {
	return "/" + apiVer + "/" + p.path + "/" + issue + "/" + p.role
}

func (p *pki) getRevokeURL() string {
	return "/" + apiVer + "/" + p.path + "/" + revoke
}
