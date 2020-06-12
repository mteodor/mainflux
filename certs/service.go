// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
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

	// ErrExternalKeyNotFound indicates a non-existent bootstrap configuration for given external key
	ErrExternalKeyNotFound = errors.New("failed to get bootstrap configuration for given external key")

	// ErrSecureBootstrap indicates erron in getting bootstrap configuration for given encrypted external key
	ErrSecureBootstrap = errors.New("failed to get bootstrap configuration for given encrypted external key")
)

var _ Service = (*certsService)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	//
	IssueCert(token string)
}

type certsService struct {
	auth  mainflux.AuthNServiceClient
	certs CertsRepository
	sdk   mfsdk.SDK
}

// New returns new Bootstrap service.
func New(auth mainflux.AuthNServiceClient, certs CertsRepository, sdk mfsdk.SDK) Service {
	return &certsService{
		certs: certs,
		sdk:   sdk,
		auth:  auth,
	}
}

func (cs *certsService) IssueCert(token string) {
}
