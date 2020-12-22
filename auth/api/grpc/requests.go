// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/authz"
)

type identityReq struct {
	token string
	kind  uint32
}

func (req identityReq) validate() error {
	if req.token == "" {
		return auth.ErrMalformedEntity
	}
	if req.kind != auth.UserKey &&
		req.kind != auth.APIKey &&
		req.kind != auth.RecoveryKey {
		return auth.ErrMalformedEntity
	}

	return nil
}

type issueReq struct {
	id      string
	email   string
	keyType uint32
}

func (req issueReq) validate() error {
	if req.email == "" {
		return auth.ErrUnauthorizedAccess
	}
	if req.keyType != auth.UserKey &&
		req.keyType != auth.APIKey &&
		req.keyType != auth.RecoveryKey {
		return auth.ErrMalformedEntity
	}

	return nil
}

type assignReq struct {
	token    string
	groupID  string
	memberID string
}

func (req assignReq) validate() error {
	if req.token == "" {
		return auth.ErrUnauthorizedAccess
	}
	if req.groupID == "" || req.memberID == "" {
		return auth.ErrMalformedEntity
	}
	return nil
}

type membersReq struct {
	token   string
	groupID string
}

func (req membersReq) validate() error {
	if req.token == "" {
		return auth.ErrUnauthorizedAccess
	}
	if req.groupID == "" {
		return auth.ErrMalformedEntity
	}
	return nil
}

// AuthZReq represents authorization request. It contains:
// 1. subject - an action invoker
// 2. object - an entity over which action will be executed
// 3. action - type of action that will be executed (read/write)
type AuthZReq struct {
	Sub string
	Obj string
	Act string
}

func (req AuthZReq) validate() error {
	if req.Sub == "" {
		return authz.ErrInvalidReq
	}

	if req.Obj == "" {
		return authz.ErrInvalidReq
	}

	if req.Act == "" {
		return authz.ErrInvalidReq
	}

	return nil
}

type AssignmentReq struct {
	GroupID  string
	MemberID string
}

func (req AssignmentReq) validate() error {
	if req.GroupID == "" {
		return authz.ErrInvalidReq
	}

	if req.MemberID == "" {
		return authz.ErrInvalidReq
	}

	return nil
}
