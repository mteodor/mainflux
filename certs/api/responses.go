// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
)

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

type certsPageRes struct {
	pageRes
	Certs []thingCertsRes `json:"certs"`
	Error string          `json:"error,omitempty"`
}

type thingCertsRes struct {
	ID     string            `json:"thing_id"`
	Cert   map[string]string `json:"thing_cert"`
	Key    map[string]string `json:"thing_key"`
	Serial string            `json:"serial"`
	CACert string            `json:"ca_cert"`
	Error  string            `json:"error"`
}

func (res certsPageRes) Code() int {
	return http.StatusCreated
}

func (res certsPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res certsPageRes) Empty() bool {
	return false
}

func (res thingCertsRes) Code() int {
	return http.StatusCreated
}

func (res thingCertsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res thingCertsRes) Empty() bool {
	return false
}
