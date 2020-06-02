package provision

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
	"math/big"
	"os"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	SDK "github.com/mainflux/mainflux/pkg/sdk/go"
)

const (
	externalIDKey = "external_id"
	gateway       = "gateway"
	Active        = 1

	control = "control"
	data    = "data"
	export  = "export"
)

var (
	ErrFailedToCreateToken      = errors.New("failed to create access token")
	ErrEmptyThingsList          = errors.New("things list in configuration empty")
	ErrEmptyChannelsList        = errors.New("channels list in configuration is empty")
	ErrFailedChannelCreation    = errors.New("failed to create channel")
	ErrFailedChannelRetrieval   = errors.New("failed to retrieve channel")
	ErrFailedThingCreation      = errors.New("failed to create thing")
	ErrFailedThingRetrieval     = errors.New("failed to retrieve thing")
	ErrMissingCredentials       = errors.New("missing credentials")
	ErrFailedBootstrapRetrieval = errors.New("failed to retrieve bootstrap")
	ErrFailedCertCreation       = errors.New("failed to create certificates")
	ErrFailedBootstrap          = errors.New("failed to create bootstrap config")
	ErrGatewayUpdate            = errors.New("failed to updated gateway metadata")

	errFailedCertCreation     = errors.New("failed creating certificate")
	errFailedDateSetting      = errors.New("failed setting date")
	errFailedPemDataWrite     = errors.New("failed writing pem data")
	errFailedPemKeyWrite      = errors.New("failed writing pem key data")
	errFailedSerialGeneration = errors.New("failed generating certificates serial")
)

var _ Service = (*provisionService)(nil)

// Service specifies Provision service API.
type Service interface {
	// Provision is the only method this API specifies. Depending on the configuration,
	// the following actions will can be executed:
	// - create a Thing based on external_id (eg. MAC address)
	// - create multiple Channels
	// - create Bootstrap configuration
	// - whitelist Thing in Bootstrap configuration == connect Thing to Channels
	Provision(token, name, externalID, externalKey string) (Result, error)

	// Certs creates certificate for things that communicate over mTLS
	// A duration string is a possibly signed sequence of decimal numbers,
	// each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// rsaBits for certificate key
	Certs(token, thingId, duration string, rsaBits int) (string, string, error)
}

type provisionService struct {
	logger logger.Logger
	sdk    SDK.SDK
	conf   Config
}

// Result represent what is created with additional info.
type Result struct {
	Things      []SDK.Thing       `json:"things,omitempty"`
	Channels    []SDK.Channel     `json:"channels,omitempty"`
	ClientCert  map[string]string `json:"client_cert,omitempty"`
	ClientKey   map[string]string `json:"client_key,omitempty"`
	CACert      string            `json:"ca_cert,omitempty"`
	Whitelisted map[string]bool   `json:"whitelisted,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// New returns new provision service.
func New(cfg Config, sdk SDK.SDK, logger logger.Logger) Service {
	return &provisionService{
		logger: logger,
		conf:   cfg,
		sdk:    sdk,
	}
}

// Provision is provision method for creating setup according to
// provision layout specified in config.toml
func (ps *provisionService) Provision(name, token, externalID, externalKey string) (res Result, err error) {
	var channels []SDK.Channel
	var things []SDK.Thing
	defer ps.recover(&err, &things, &channels, &token)

	token, err = ps.createIfNotValidToken(token)
	if err != nil {
		return res, err
	}

	if len(ps.conf.Things) == 0 {
		return res, ErrEmptyThingsList
	}
	if len(ps.conf.Channels) == 0 {
		return res, ErrEmptyChannelsList
	}
	for _, thing := range ps.conf.Things {
		// If thing in configs contains metadata with external_id
		// set value for it from the provision request
		if _, ok := thing.Metadata[externalIDKey]; ok {
			thing.Metadata[externalIDKey] = externalID
		}

		th := SDK.Thing{
			Metadata: thing.Metadata,
		}
		if name == "" {
			name = thing.Name
		}
		th.Name = name
		thID, err := ps.sdk.CreateThing(th, token)
		if err != nil {
			res.Error = err.Error()
			return res, errors.Wrap(ErrFailedThingCreation, err)
		}
		// Get newly created thing (in order to get the key).
		th, err = ps.sdk.Thing(thID, token)
		if err != nil {
			e := errors.Wrap(err, fmt.Errorf("thing id: %s", thID))
			return res, errors.Wrap(ErrFailedThingRetrieval, e)
		}
		things = append(things, th)
	}

	for _, channel := range ps.conf.Channels {
		ch := SDK.Channel{
			Name:     channel.Name,
			Metadata: channel.Metadata,
		}
		chCreated, err := ps.sdk.CreateChannel(ch, token)
		if err != nil {
			return res, err
		}
		ch, err = ps.sdk.Channel(chCreated, token)
		if err != nil {
			e := errors.Wrap(err, fmt.Errorf("channel id: %s", chCreated))
			return res, errors.Wrap(ErrFailedChannelRetrieval, e)
		}
		channels = append(channels, ch)
	}

	res = Result{
		Things:      things,
		Channels:    channels,
		Whitelisted: map[string]bool{},
		ClientCert:  map[string]string{},
		ClientKey:   map[string]string{},
	}

	var cert SDK.Cert
	var bs SDK.BootstrapConfig
	for _, thing := range things {
		bootstrap := false
		if _, ok := thing.Metadata[externalIDKey]; ok {
			bootstrap = true
		}
		var chanIDs []string
		for _, ch := range channels {
			chanIDs = append(chanIDs, ch.ID)
		}
		if ps.conf.Bootstrap.Provision && bootstrap {
			bsReq := SDK.BootstrapConfig{
				ThingID:     thing.ID,
				ExternalID:  externalID,
				ExternalKey: externalKey,
				Channels:    chanIDs,
				CACert:      res.CACert,
				ClientCert:  cert.ClientCert,
				ClientKey:   cert.ClientKey,
				Content:     ps.conf.Bootstrap.Content,
			}
			bsid, err := ps.sdk.AddBootstrap(token, bsReq)
			if err != nil {
				return Result{}, errors.Wrap(ErrFailedBootstrap, err)
			}

			bs, err = ps.sdk.ViewBootstrap(token, bsid)
			if err != nil {
				return Result{}, err
			}
		}

		if ps.conf.Bootstrap.X509Provision {
			var cert SDK.Cert
			if ps.conf.Server.MfCertsURL == "" {

				cert.ClientCert, cert.ClientKey, err = ps.certs(thing.Key, ps.conf.Certs.DaysValid, ps.conf.Certs.RsaBits)
				if err != nil {
					e := errors.Wrap(err, fmt.Errorf("thing id: %s", thing.ID))
					return res, errors.Wrap(ErrFailedCertCreation, e)
				}
			} else {
				cert, err = ps.sdk.Cert(thing.ID, thing.Key, token)
				if err != nil {
					e := errors.Wrap(err, fmt.Errorf("thing id: %s", thing.ID))
					return res, errors.Wrap(ErrFailedCertCreation, e)
				}
			}

			res.ClientCert[thing.ID] = cert.ClientCert
			res.ClientKey[thing.ID] = cert.ClientKey
			res.CACert = cert.CACert
		}

		if ps.conf.Bootstrap.AutoWhiteList {
			wlReq := SDK.BootstrapConfig{
				MFThing: thing.ID,
				State:   Active,
			}
			if err := ps.sdk.Whitelist(token, wlReq); err != nil {
				res.Error = err.Error()
				return res, SDK.ErrFailedWhitelist
			}
			res.Whitelisted[thing.ID] = true
		}

		if ps.conf.Server.MfCertsURL == "" && ps.conf.Bootstrap.X509Provision == true {

		}

	}

	ps.updateGateway(token, bs, channels)
	return res, nil
}

func (ps *provisionService) createIfNotValidToken(token string) (string, error) {
	if token != "" {
		return token, nil
	}

	// If no token in request is provided
	// use API key provided in config file or env
	if ps.conf.Server.MfAPIKey != "" {
		return ps.conf.Server.MfAPIKey, nil
	}

	// If no API key use username and password provided to create access token.
	if ps.conf.Server.MfUser == "" || ps.conf.Server.MfPass == "" {
		return token, ErrMissingCredentials
	}

	u := SDK.User{
		Email:    ps.conf.Server.MfUser,
		Password: ps.conf.Server.MfPass,
	}
	token, err := ps.sdk.CreateToken(u)
	if err != nil {
		return token, errors.Wrap(ErrFailedToCreateToken, err)
	}

	return token, nil
}

func (ps *provisionService) Certs(token, thingId, daysValid string, rsaBits int) (string, string, error) {
	token, err := ps.createIfNotValidToken(token)
	if err != nil {
		return "", "", err
	}

	th, err := ps.sdk.Thing(thingId, token)
	if err != nil {
		return "", "", errors.Wrap(SDK.ErrUnauthorized, err)
	}

	return ps.certs(th.Key, daysValid, rsaBits)
}

func (ps *provisionService) certs(thingKey, daysValid string, rsaBits int) (string, string, error) {
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

	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, ps.conf.Certs.CA, publicKey(priv), ps.conf.Certs.Cert.PrivateKey)
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

	if err := pem.Encode(buffKeyOut, pemBlockForKey(priv)); err != nil {
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

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func (ps *provisionService) updateGateway(token string, bs SDK.BootstrapConfig, channels []SDK.Channel) error {
	var gw Gateway
	for _, ch := range channels {
		switch ch.Metadata["type"] {
		case control:
			gw.CtrlChannelID = ch.ID
		case data:
			gw.DataChannelID = ch.ID
		case export:
			gw.ExportChannelID = ch.ID
		}
	}
	gw.ExternalID = bs.ExternalID
	gw.ExternalKey = bs.ExternalKey
	gw.CfgID = bs.MFThing
	gw.Type = gateway

	th, err := ps.sdk.Thing(bs.MFThing, token)
	if err != nil {
		return errors.Wrap(ErrGatewayUpdate, err)
	}
	b, err := json.Marshal(gw)
	if err != nil {
		return errors.Wrap(ErrGatewayUpdate, err)
	}
	if err := json.Unmarshal(b, &th.Metadata); err != nil {
		return errors.Wrap(ErrGatewayUpdate, err)
	}
	if err := ps.sdk.UpdateThing(th, token); err != nil {
		return errors.Wrap(ErrGatewayUpdate, err)
	}
	return nil
}

func (ps *provisionService) errLog(err error) {
	if err != nil {
		ps.logger.Error(fmt.Sprintf("Error recovering: %s", err))
	}
}

func clean(ps *provisionService, things []SDK.Thing, channels []SDK.Channel, token string) {
	for _, t := range things {
		ps.errLog(ps.sdk.DeleteThing(t.ID, token))
	}
	for _, c := range channels {
		ps.errLog(ps.sdk.DeleteThing(c.ID, token))
	}
}

func (ps *provisionService) recover(e *error, ths *[]SDK.Thing, chs *[]SDK.Channel, tkn *string) {
	things, channels, token, err := *ths, *chs, *tkn, *e
	if e == nil {
		return
	}
	if errors.Contains(err, ErrFailedThingRetrieval) || errors.Contains(err, ErrFailedChannelCreation) {
		for _, th := range things {
			ps.errLog(ps.sdk.DeleteThing(th.ID, token))
		}
		return
	}

	if errors.Contains(err, ErrFailedChannelRetrieval) || errors.Contains(err, ErrFailedCertCreation) {
		for _, th := range things {
			ps.errLog(ps.sdk.DeleteThing(th.ID, token))
		}
		for _, ch := range channels {
			ps.errLog(ps.sdk.DeleteChannel(ch.ID, token))
		}
		return
	}

	if errors.Contains(err, ErrFailedBootstrap) {
		clean(ps, things, channels, token)
		if ps.conf.Bootstrap.X509Provision {
			for _, th := range things {
				ps.errLog(ps.sdk.RemoveCert(th.ID, token))
			}
		}
		return
	}

	if errors.Contains(err, SDK.ErrFailedWhitelist) {
		clean(ps, things, channels, token)
		for _, th := range things {
			if ps.conf.Bootstrap.X509Provision {
				ps.errLog(ps.sdk.RemoveCert(th.ID, token))
			}
			bs, err := ps.sdk.ViewBootstrap(token, th.ID)
			ps.errLog(errors.Wrap(ErrFailedBootstrapRetrieval, err))
			ps.errLog(ps.sdk.RemoveBootstrap(token, bs.MFThing))
		}
		return
	}

	if errors.Contains(err, ErrGatewayUpdate) {
		clean(ps, things, channels, token)
		for _, th := range things {
			if ps.conf.Bootstrap.X509Provision {
				ps.errLog(ps.sdk.RemoveCert(th.ID, token))
			}
			bs, err := ps.sdk.ViewBootstrap(token, th.ID)
			ps.errLog(errors.Wrap(ErrFailedBootstrapRetrieval, err))
			ps.errLog(ps.sdk.RemoveBootstrap(token, bs.MFThing))
		}
		return
	}

}
