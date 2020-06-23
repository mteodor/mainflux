package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/certs"
)

func doIssueCert(svc certs.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {

		req := request.(addCertsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		token := req.token

		res, err := svc.IssueCert(req.ThingID, req.Valid, req.RsaBits, token)

		if err != nil {
			return certsResponse{Error: err.Error()}, nil
		}

		return res, nil

	}
}
