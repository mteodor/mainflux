// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/users"
)

func registrationEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(userReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.Register(ctx, req.user); err != nil {
			return tokenRes{}, err
		}
		return tokenRes{}, nil
	}
}

// Password reset request endpoint.
// When successful password reset link is generated.
// Link is generated using MF_TOKEN_RESET_ENDPOINT env.
// and value from Referer header for host.
// {Referer}+{MF_TOKEN_RESET_ENDPOINT}+{token=TOKEN}
// http://mainflux.com/reset-request?token=xxxxxxxxxxx.
// Email with a link is being sent to the user.
// When user clicks on a link it should get the ui with form to
// enter new password, when form is submitted token and new password
// must be sent as PUT request to 'password/reset' passwordResetEndpoint
func passwordResetRequestEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(passwResetReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		res := passwChangeRes{}
		email := req.Email

		if err := svc.GenerateResetToken(ctx, email, req.Host); err != nil {
			return nil, err
		}
		res.Msg = MailSent
		return res, nil
	}
}

// This is endpoint that actually sets new password in password reset flow.
// When user clicks on a link in email finally ends on this endpoint as explained in
// the comment above.
func passwordResetEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetTokenReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res := passwChangeRes{}

		if err := svc.ResetPassword(ctx, req.Token, req.Password); err != nil {
			return nil, err
		}
		res.Msg = ""
		return res, nil
	}
}

func viewUserEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewUserReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		u, err := svc.ViewUser(ctx, req.token)
		if err != nil {
			return nil, err
		}

		return viewUserRes{
			ID:       u.ID,
			Email:    u.Email,
			Metadata: u.Metadata,
		}, nil
	}
}

func updateUserEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateUserReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		user := users.User{
			Metadata: req.Metadata,
		}
		err := svc.UpdateUser(ctx, req.token, user)
		if err != nil {
			return nil, err
		}

		return updateUserRes{}, nil
	}
}

func passwordChangeEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(passwChangeReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res := passwChangeRes{}

		if err := svc.ChangePassword(ctx, req.Token, req.Password, req.OldPassword); err != nil {
			return nil, err
		}

		return res, nil
	}
}

func loginEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(userReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		token, err := svc.Login(ctx, req.user)
		if err != nil {
			return nil, err
		}

		return tokenRes{token}, nil
	}
}


func createGroupEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createThingReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		th := things.Thing{
			Key:      req.Key,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		saved, err := svc.CreateThings(ctx, req.token, th)
		if err != nil {
			return nil, err
		}

		res := thingRes{
			ID:      saved[0].ID,
			created: true,
		}

		return res, nil
	}
}


func updateGroupEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateGroupReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		thing := things.Group{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}

		if err := svc.UpdateGroup(ctx, req.token, thing); err != nil {
			return nil, err
		}

		res := thingRes{ID: req.id, created: false}
		return res, nil
	}
}

func updateKeyEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateKeyReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.UpdateKey(ctx, req.token, req.id, req.Key); err != nil {
			return nil, err
		}

		res := thingRes{ID: req.id, created: false}
		return res, nil
	}
}

func viewGroupEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		thing, err := svc.ViewGroup(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}

		res := viewGroupRes{
			ID:       thing.ID,
			Owner:    thing.Owner,
			Name:     thing.Name,
			Key:      thing.Key,
			Metadata: thing.Metadata,
		}
		return res, nil
	}
}

func listGroupsEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listResourcesReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListGroups(ctx, req.token, req.offset, req.limit, req.name, req.metadata)
		if err != nil {
			return nil, err
		}

		res := thingsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Groups: []viewGroupRes{},
		}
		for _, thing := range page.Groups {
			view := viewGroupRes{
				ID:       thing.ID,
				Owner:    thing.Owner,
				Name:     thing.Name,
				Key:      thing.Key,
				Metadata: thing.Metadata,
			}
			res.Groups = append(res.Groups, view)
		}

		return res, nil
	}
}

func removeThingEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		err := req.validate()
		if err == things.ErrNotFound {
			return removeRes{}, nil
		}

		if err != nil {
			return nil, err
		}

		if err := svc.RemoveThing(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return removeRes{}, nil
	}
}
