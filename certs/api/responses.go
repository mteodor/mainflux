package api

import (
	"net/http"
)

type certsResponse struct {
	ClientCert map[string]string `json:"client_cert,omitempty"`
	ClientKey  map[string]string `json:"client_key,omitempty"`
	CACert     string            `json:"ca_cert,omitempty"`
}

func (res certsResponse) Code() int {
	return http.StatusCreated
}

func (res certsResponse) Headers() map[string]string {
	return map[string]string{}
}

func (res certsResponse) Empty() bool {
	return false
}
