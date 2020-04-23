package certs_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux/provision/certs"
	SDK "github.com/mainflux/mainflux/sdk/go"
	"github.com/stretchr/testify/assert"
)

const (
	contentType = "application/json"
	invalid     = "invalid"
	exists      = "exists"
	valid       = "valid"
)

type handler func(http.ResponseWriter, *http.Request)

func (h handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h(rw, r)
}

type tokenRes struct {
	Token string `json:"token,omitempty"`
}

type certReq struct {
	ThingID  string `json:"id,omitempty"`
	ThingKey string `json:"key,omitempty"`
}

func auth(rw http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") != valid {
		rw.WriteHeader(http.StatusForbidden)
		return false
	}
	return true
}

func ct(rw http.ResponseWriter, r *http.Request) bool {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		rw.WriteHeader(http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func delete(rw http.ResponseWriter, r *http.Request) {
	if !auth(rw, r) {
		return
	}
	id := bone.GetValue(r, "id")
	if id == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}

func createToken(rw http.ResponseWriter, r *http.Request) {
	var u SDK.User
	if !ct(rw, r) {
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if u.Email == invalid || u.Password == invalid {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	data, _ := json.Marshal(tokenRes{valid})
	rw.Write(data)
}

func cert(rw http.ResponseWriter, r *http.Request) {
	if !auth(rw, r) || !ct(rw, r) {
		return
	}
	crt := certReq{}
	if err := json.NewDecoder(r.Body).Decode(&crt); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if crt.ThingID == "" || crt.ThingKey == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if crt.ThingID == exists || crt.ThingKey == exists {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	data, _ := json.Marshal(certs.Cert{})
	rw.WriteHeader(http.StatusCreated)
	rw.Write(data)
}

func saveToBootstrap(rw http.ResponseWriter, r *http.Request) {
	if !auth(rw, r) || !ct(rw, r) {
		return
	}
	var cfg SDK.BoostrapConfig
	json.NewDecoder(r.Body).Decode(&cfg)
	defer r.Body.Close()

	if cfg.ThingID == exists {
		rw.WriteHeader(http.StatusConflict)
		return
	}
	if cfg.Channels[0] == invalid || cfg.Channels[1] == invalid {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func whitelist(rw http.ResponseWriter, r *http.Request) {
	if !auth(rw, r) || !ct(rw, r) {
		return
	}
	id := bone.GetValue(r, "id")
	if id == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if id != exists {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	d := make(map[string]int)
	json.NewDecoder(r.Body).Decode(&d)
	defer r.Body.Close()

	if s, ok := d["state"]; ok {
		if s != 0 && s != 1 {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		rw.WriteHeader(http.StatusOK)
		return
	}
	rw.WriteHeader(http.StatusBadRequest)
}

func removeCert(rw http.ResponseWriter, r *http.Request) {}

func newSDK() certs.SDK {
	r := bone.New()
	r.Post("/tokens", handler(createToken))
	r.Post("/certs", handler(cert))
	r.Delete("/certs/:id", handler(delete))
	svc := httptest.NewServer(r)
	crt := fmt.Sprintf("%s/certs", svc.URL)
	return certs.New(crt)
}

func TestCert(t *testing.T) {
	sdk := newSDK()

	cases := []struct {
		desc  string
		id    string
		key   string
		token string
		err   error
	}{
		{
			desc:  "Create cert successfully",
			id:    valid,
			key:   valid,
			token: valid,
			err:   nil,
		},
		{
			desc:  "Create cert unauthorized",
			id:    valid,
			key:   valid,
			token: invalid,
			err:   certs.ErrCerts,
		},
		{
			desc:  "Create cert with an existing id",
			id:    exists,
			key:   valid,
			token: valid,
			err:   certs.ErrCerts,
		},
		{
			desc:  "Create cert with an existing key",
			id:    valid,
			key:   exists,
			token: valid,
			err:   certs.ErrCerts,
		},
	}
	for _, tc := range cases {
		_, err := sdk.Cert(tc.id, tc.key, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestRemoveCert(t *testing.T) {
	sdk := newSDK()

	cases := []struct {
		desc  string
		key   string
		token string
		err   error
	}{
		{
			desc:  "Delete cert successfully",
			key:   valid,
			token: valid,
			err:   nil,
		},
		{
			desc:  "Delete cert unauthorized",
			key:   valid,
			token: invalid,
			err:   certs.ErrUnauthorized,
		},
		{
			desc:  "Delete cert wrong ID",
			key:   "",
			token: valid,
			err:   certs.ErrCertsRemove,
		},
	}
	for _, tc := range cases {
		err := sdk.RemoveCert(tc.key, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}
