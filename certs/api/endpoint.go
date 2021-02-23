// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/certs"
)

func issueCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addCertsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res, err := svc.IssueCert(ctx, req.token, req.ThingID, req.Valid, req.KeyBits, req.KeyType)
		if err != nil {
			return thingCertsRes{Error: err.Error()}, nil
		}
		return thingCertsRes{
			Serial: res.Serial,
			ID:     res.ThingID,
			Key:    map[string]string{res.Serial: res.ClientKey},
			Cert:   map[string]string{res.Serial: res.ClientCert},
			CACert: res.IssuingCA,
		}, nil
	}
}

func listCerts(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListCerts(ctx, req.token, req.thingID, req.offset, req.limit)
		if err != nil {
			return certsPageRes{
				Error: err.Error(),
			}, err
		}
		res := certsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Certs: []thingCertsRes{},
		}

		for _, cert := range page.Certs {
			view := thingCertsRes{
				Serial: cert.Serial,
				ID:     cert.ThingID,
				Key:    map[string]string{cert.Serial: cert.ClientKey},
				Cert:   map[string]string{cert.Serial: cert.ClientCert},
				CACert: cert.IssuingCA,
			}
			res.Certs = append(res.Certs, view)
		}
		return res, nil
	}
}

func revokeCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(revokeReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		return svc.RevokeCert(ctx, req.token, req.thingID)
	}
}
