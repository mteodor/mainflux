// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/authz"
)

func MakeAuthorizeEndpoint(svc authz.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AuthZReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		p := authz.Policy{
			Subject: req.Sub,
			Object:  req.Obj,
			Action:  req.Act,
		}

		err := svc.Authorize(ctx, p)
		return ErrorRes{Err: err}, nil
	}
}
