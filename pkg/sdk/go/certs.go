package sdk

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/mainflux/mainflux/errors"
)

var (
	ErrCertsCreation = errors.New("failed to create certificate")

	errFailedCertCreation     = errors.New("failed creating certificate")
	errFailedDateSetting      = errors.New("failed setting date")
	errFailedPemDataWrite     = errors.New("failed writing pem data")
	errFailedPemKeyWrite      = errors.New("failed writing pem key data")
	errFailedSerialGeneration = errors.New("failed generating certificates serial")
	errFailedCertLoading      = errors.New("failed to load certificate")
	errFailedCertDecode       = errors.New("failed to decode certificate")
)

// Cert represents certs data.
type Cert struct {
	CACert     string `json:"ca_cert,omitempty"`
	ClientKey  string `json:"client_key,omitempty"`
	ClientCert string `json:"client_cert,omitempty"`
}

func (sdk mfSDK) Cert(thingID, daysValid string, rsaBits int, token string) (Cert, error) {
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

	if sdk.certsURL == "" {
		c.ClientCert, c.ClientKey, err = sdk.certs(th.Key, daysValid, rsaBits)
		if err != nil {
			return Cert{}, errors.Wrap(ErrCertsCreation, err)
		}
		return c, err
	}

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
		println(err.Error())
		return Cert{}, err
	}
	if err := json.Unmarshal(body, &c); err != nil {
		return Cert{}, err
	}
	return c, nil
}

func (sdk mfSDK) certs(thingKey, daysValid string, rsaBits int) (string, string, error) {
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

	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, sdk.certsCA, publicKey(priv), sdk.certsCert.PrivateKey)
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

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
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

func (sdk mfSDK) RemoveCert(id, token string) error {
	res, err := request(http.MethodDelete, token, fmt.Sprintf("%s/%s", sdk.certsURL, id), nil)
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
	ThingID  string `json:"thing_id,omitempty"`
	ThingKey string `json:"thing_key,omitempty"`
}
