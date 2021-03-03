// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs_test

import (
	"strconv"
	"testing"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/things/mocks"
)

const (
	wrongID    = ""
	wrongValue = "wrong-value"
	email      = "user@example.com"
	token      = "token"
	token2     = "token2"
	thingsNum  = 5
	n          = uint64(10)
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
	pki = mocks.NewPkiAgent()
	sdk := mfsdk.NewSDK(config)
	return certs.New(auth, sdk, pki)
}

func newThingsService(auth mainflux.AuthServiceClient) things.Service {
	things := make(map[string]things.Thing, thingsNum)
	for i := 0; i < thingsNum; i++ {
		id := strconv.Itoa(i + 1)
		things[id] = things.Thing{
			ID:    id,
			Owner: email,
		}
	}

	return mocks.NewThingsService(things, map[string]things.Channels{}, auth)
}

func TestIssueCert(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{validToken: email})
	server := newThingsServer(newThingsService(users))
	svc := newService(map[string]string{token: email}, server.URL)
}
