// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/users"
	"google.golang.org/grpc"
)

var _ mainflux.AuthServiceClient = (*authNServiceMock)(nil)

type authNServiceMock struct {
	users map[string]string
}

// NewAuthService creates mock of users service.
func NewAuthService(users map[string]string) mainflux.AuthServiceClient {
	return &authNServiceMock{users}
}

func (svc authNServiceMock) Identify(ctx context.Context, in *mainflux.Token, opts ...grpc.CallOption) (*mainflux.UserIdentity, error) {
	if id, ok := svc.users[in.Value]; ok {
		return &mainflux.UserIdentity{Id: id, Email: id}, nil
	}
	return nil, users.ErrUnauthorizedAccess
}

func (svc authNServiceMock) Issue(ctx context.Context, in *mainflux.IssueReq, opts ...grpc.CallOption) (*mainflux.Token, error) {
	if id, ok := svc.users[in.GetEmail()]; ok {
		switch in.Type {
		default:
			return &mainflux.Token{Value: id}, nil
		}
	}
	return nil, users.ErrUnauthorizedAccess
}

func (svc serviceMock) Authorize(ctx context.Context, req *mainflux.AuthorizeReq, _ ...grpc.CallOption) (r *mainflux.AuthorizeRes, err error) {
	panic("not implemented")
}

func (svc serviceMock) Members(ctx context.Context, req *mainflux.MembersReq, _ ...grpc.CallOption) (r *mainflux.MembersRes, err error) {
	panic("not implemented")
}

func (svc serviceMock) Assign(ctx context.Context, req *mainflux.Assignment, _ ...grpc.CallOption) (r *empty.Empty, err error) {
	panic("not implemented")
}
