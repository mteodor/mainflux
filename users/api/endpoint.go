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
		req := request.(createGroupReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		group := users.Group{
			Name:        req.Name,
			Description: req.Description,
		}
		saved, err := svc.CreateGroup(ctx, req.token, group)
		if err != nil {
			return nil, err
		}

		res := groupRes{
			Name:    saved.Name,
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

		group := users.Group{
			Name:        req.Name,
			Description: req.Description,
			Metadata:    req.Metadata,
		}

		if err := svc.UpdateGroup(ctx, req.token, group); err != nil {
			return nil, err
		}

		res := groupRes{
			Name:        group.Name,
			Description: group.Description,
			Metadata:    group.Metadata,
			created:     false,
		}
		return res, nil
	}
}

func viewGroupEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewGroupReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		group, err := svc.ViewGroup(ctx, req.token, req.Name)
		if err != nil {
			return nil, err
		}

		res := viewGroupRes{
			Name:        group.Name,
			Description: group.Description,
			Metadata:    group.Metadata,
		}
		return res, nil
	}
}

func listGroupsEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListGroups(ctx, req.token, req.offset, req.limit, req.name, req.metadata)
		if err != nil {
			return nil, err
		}

		res := groupPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Groups: []viewGroupRes{},
		}
		for _, group := range page.Groups {
			view := viewGroupRes{
				Name:        group.Name,
				Description: group.Description,
				Metadata:    group.Metadata,
			}
			res.Groups = append(res.Groups, view)
		}

		return res, nil
	}
}
