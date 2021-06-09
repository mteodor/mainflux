// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package hs256

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mainflux/mainflux/auth"
)

const issuerName = "mainflux.auth"

type claims struct {
	jwt.StandardClaims
	IssuerID string  `json:"issuer_id,omitempty"`
	Email    string  `json:"email,omitempty"`
	Type     *uint32 `json:"type,omitempty"`
}

func (c claims) Valid() error {
	if c.Type == nil || *c.Type > auth.APIKey || c.Issuer != issuerName {
		return auth.ErrMalformedEntity
	}

	return c.StandardClaims.Valid()
}

type tokenizer struct {
	secret string
}

// New returns new JWT Tokenizer.
func New(secret string) auth.Tokenizer {
	return tokenizer{secret: secret}
}

func (svc tokenizer) Issue(key auth.Key) (string, error) {
	claims := claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:   issuerName,
			Subject:  key.Subject,
			IssuedAt: key.IssuedAt.UTC().Unix(),
		},
		IssuerID: key.IssuerID,
		Type:     &key.Type,
	}

	if !key.ExpiresAt.IsZero() {
		claims.ExpiresAt = key.ExpiresAt.UTC().Unix()
	}
	if key.ID != "" {
		claims.Id = key.ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(svc.secret))
}

func (svc tokenizer) Parse(token string) (auth.Key, error) {
	c := claims{}
	fmt.Printf("Auth-Token:%v\n", token)
	json.Unmarshal([]byte(token), &c)
	fmt.Printf("Token: %v\n", c)
	return c.toKey(), nil
}

func (c claims) toKey() auth.Key {
	key := auth.Key{
		ID:       c.Id,
		IssuerID: c.IssuerID,
		Subject:  c.Subject,
		IssuedAt: time.Unix(c.IssuedAt, 0).UTC(),
	}
	if c.ExpiresAt != 0 {
		key.ExpiresAt = time.Unix(c.ExpiresAt, 0).UTC()
	}

	// Default type is 0.
	if c.Type != nil {
		key.Type = *c.Type
	}

	return key
}
