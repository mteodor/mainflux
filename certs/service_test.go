// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/certs/mocks"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/things"
	httpapi "github.com/mainflux/mainflux/things/api/things/http"
	"github.com/opentracing/opentracing-go/mocktracer"
)

const (
	wrongID    = ""
	wrongValue = "wrong-value"
	email      = "user@example.com"
	token      = "token"
	token2     = "token2"
	thingsNum  = 5
	n          = uint64(10)

	cfgLogLevel     = "error"
	cfgClientTLS    = false
	cfgCACerts      = ""
	cfgPort         = "8204"
	cfgServerCert   = ""
	cfgServerKey    = ""
	cfgBaseURL      = "http://localhost"
	cfgThingsPrefix = ""
	cfgJaegerURL    = ""
	cfgAuthURL      = "localhost:8181"

	caPath            = "docker/ssl/certs/ca.crt"
	caKeyPath         = "docker/ssl/certs/ca.key"
	cfgSignHoursValid = "24h"
	cfgSignRSABits    = 2048
)

// // Service specifies an API that must be fulfilled by the domain service
// // implementation, and all of its decorators (e.g. logging & metrics).
// type Service interface {
// 	// IssueCert issues certificate for given thing id if access is granted with token
// 	IssueCert(ctx context.Context, token, thingID, daysValid string, keyBits int, keyType string) (Cert, error)

// 	// ListCerts lists all certificates issued for given owner
// 	ListCerts(ctx context.Context, token, thingID string, offset, limit uint64) (Page, error)

// 	// RevokeCert revokes certificate for given thing
// 	RevokeCert(ctx context.Context, token, thingID string) (Revoke, error)
// }
func newService(tokens map[string]string, url string) certs.Service {
	auth := mocks.NewAuthService(tokens)
	config := mfsdk.Config{
		BaseURL: url,
	}
	pki := mocks.NewPkiAgent()
	sdk := mfsdk.NewSDK(config)
	repo := mocks.NewCertsRepository()

	tlsCert, caCert, _ := loadCertificates(caPath, caKeyPath)

	c := certs.Config{
		LogLevel:       cfgLogLevel,
		ClientTLS:      cfgClientTLS,
		ServerCert:     cfgServerCert,
		ServerKey:      cfgServerKey,
		BaseURL:        cfgBaseURL,
		ThingsPrefix:   cfgThingsPrefix,
		JaegerURL:      cfgJaegerURL,
		AuthURL:        cfgAuthURL,
		SignTLSCert:    tlsCert,
		SignX509Cert:   caCert,
		SignHoursValid: cfgSignHoursValid,
		SignRSABits:    cfgSignRSABits,
	}

	return certs.New(auth, repo, sdk, c, pki)
}

func newThingsService(auth mainflux.AuthServiceClient) things.Service {
	ths := make(map[string]things.Thing, thingsNum)
	for i := 0; i < thingsNum; i++ {
		id := strconv.Itoa(i + 1)
		ths[id] = things.Thing{
			ID:    id,
			Owner: email,
		}
	}

	return mocks.NewThingsService(ths, map[string]things.Channel{}, auth)
}

func TestIssueCert(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{token: email})
	server := newThingsServer(newThingsService(users))
	svc := newService(map[string]string{token: email}, server.URL)

	c, err := svc.IssueCert(context.Background(), token, "1", "2048", 2048, "rsa")
	
}

func newThingsServer(svc things.Service) *httptest.Server {
	mux := httpapi.MakeHandler(mocktracer.New(), svc)
	return httptest.NewServer(mux)
}

func loadCertificates(caPath, caKeyPath string) (tls.Certificate, *x509.Certificate, error) {
	var tlsCert tls.Certificate
	var caCert *x509.Certificate

	if caPath == "" || caKeyPath == "" {
		return tlsCert, caCert, nil
	}

	if _, err := os.Stat(caPath); os.IsNotExist(err) {
		return tlsCert, caCert, err
	}

	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		return tlsCert, caCert, err
	}

	tlsCert, err := tls.LoadX509KeyPair(caPath, caKeyPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(err, err)
	}

	b, err := ioutil.ReadFile(caPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(err, err)
	}

	block, _ := pem.Decode(b)
	if block == nil {
		log.Fatalf("No PEM data found, failed to decode CA")
	}

	caCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(err, err)
	}

	return tlsCert, caCert, nil
}
