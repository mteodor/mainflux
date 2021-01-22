// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/auth/groups"
	"github.com/mainflux/mainflux/auth/jwt"
	"github.com/mainflux/mainflux/auth/mocks"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	secret    = "secret"
	email     = "test@example.com"
	id        = "testID"
	token     = "token"
	groupName = "mfx"
)

func newService() auth.Service {
	repo := mocks.NewKeyRepository()
	groupRepo := mocks.NewGroupRepository()
	idProvider := uuid.NewMock()
	t := jwt.New(secret)
	return auth.New(repo, groupRepo, idProvider, t)
}

func TestIssue(t *testing.T) {
	svc := newService()
	_, secret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	cases := []struct {
		desc  string
		key   auth.Key
		token string
		err   error
	}{
		{
			desc: "issue user key",
			key: auth.Key{
				Type:     auth.UserKey,
				IssuedAt: time.Now(),
			},
			token: secret,
			err:   nil,
		},
		{
			desc: "issue user key with no time",
			key: auth.Key{
				Type: auth.UserKey,
			},
			token: secret,
			err:   auth.ErrInvalidKeyIssuedAt,
		},
		{
			desc: "issue API key",
			key: auth.Key{
				Type:     auth.APIKey,
				IssuedAt: time.Now(),
			},
			token: secret,
			err:   nil,
		},
		{
			desc: "issue API key unauthorized",
			key: auth.Key{
				Type:     auth.APIKey,
				IssuedAt: time.Now(),
			},
			token: "invalid",
			err:   auth.ErrUnauthorizedAccess,
		},
		{
			desc: "issue API key with no time",
			key: auth.Key{
				Type: auth.APIKey,
			},
			token: secret,
			err:   auth.ErrInvalidKeyIssuedAt,
		},
		{
			desc: "issue recovery key",
			key: auth.Key{
				Type:     auth.RecoveryKey,
				IssuedAt: time.Now(),
			},
			token: "",
			err:   nil,
		},
		{
			desc: "issue recovery with no issue time",
			key: auth.Key{
				Type: auth.RecoveryKey,
			},
			token: secret,
			err:   auth.ErrInvalidKeyIssuedAt,
		},
	}

	for _, tc := range cases {
		_, _, err := svc.Issue(context.Background(), tc.token, tc.key)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestRevoke(t *testing.T) {
	svc := newService()
	_, secret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := auth.Key{
		Type:     auth.APIKey,
		IssuedAt: time.Now(),
		IssuerID: id,
		Subject:  email,
	}
	newKey, _, err := svc.Issue(context.Background(), secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing user's key expected to succeed: %s", err))

	cases := []struct {
		desc  string
		id    string
		token string
		err   error
	}{
		{
			desc:  "revoke user key",
			id:    newKey.ID,
			token: secret,
			err:   nil,
		},
		{
			desc:  "revoke non-existing user key",
			id:    newKey.ID,
			token: secret,
			err:   nil,
		},
		{
			desc:  "revoke unauthorized",
			id:    newKey.ID,
			token: "",
			err:   auth.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		err := svc.Revoke(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestRetrieve(t *testing.T) {
	svc := newService()
	_, secret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), Subject: email, IssuerID: id})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := auth.Key{
		ID:       "id",
		Type:     auth.APIKey,
		IssuerID: id,
		Subject:  email,
		IssuedAt: time.Now(),
	}

	_, userToken, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	apiKey, apiToken, err := svc.Issue(context.Background(), secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing user's key expected to succeed: %s", err))

	_, resetToken, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.RecoveryKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing reset key expected to succeed: %s", err))

	cases := []struct {
		desc  string
		id    string
		token string
		err   error
	}{
		{
			desc:  "retrieve user key",
			id:    apiKey.ID,
			token: userToken,
			err:   nil,
		},
		{
			desc:  "retrieve non-existing user key",
			id:    "invalid",
			token: userToken,
			err:   auth.ErrNotFound,
		},
		{
			desc:  "retrieve unauthorized",
			id:    apiKey.ID,
			token: "wrong",
			err:   auth.ErrUnauthorizedAccess,
		},
		{
			desc:  "retrieve with API token",
			id:    apiKey.ID,
			token: apiToken,
			err:   auth.ErrUnauthorizedAccess,
		},
		{
			desc:  "retrieve with reset token",
			id:    apiKey.ID,
			token: resetToken,
			err:   auth.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		_, err := svc.RetrieveKey(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestIdentify(t *testing.T) {
	svc := newService()

	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	_, recoverySecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.RecoveryKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing reset key expected to succeed: %s", err))

	_, apiSecret, err := svc.Issue(context.Background(), loginSecret, auth.Key{Type: auth.APIKey, IssuerID: id, Subject: email, IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Minute)})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	exp1 := time.Now().Add(-2 * time.Second)
	_, expSecret, err := svc.Issue(context.Background(), loginSecret, auth.Key{Type: auth.APIKey, IssuedAt: time.Now(), ExpiresAt: exp1})
	assert.Nil(t, err, fmt.Sprintf("Issuing expired user key expected to succeed: %s", err))

	_, invalidSecret, err := svc.Issue(context.Background(), loginSecret, auth.Key{Type: 22, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	cases := []struct {
		desc string
		key  string
		idt  auth.Identity
		err  error
	}{
		{
			desc: "identify login key",
			key:  loginSecret,
			idt:  auth.Identity{id, email},
			err:  nil,
		},
		{
			desc: "identify recovery key",
			key:  recoverySecret,
			idt:  auth.Identity{id, email},
			err:  nil,
		},
		{
			desc: "identify API key",
			key:  apiSecret,
			idt:  auth.Identity{id, email},
			err:  nil,
		},
		{
			desc: "identify expired API key",
			key:  expSecret,
			idt:  auth.Identity{},
			err:  auth.ErrAPIKeyExpired,
		},
		{
			desc: "identify expired key",
			key:  invalidSecret,
			idt:  auth.Identity{},
			err:  auth.ErrUnauthorizedAccess,
		},
		{
			desc: "identify invalid key",
			key:  "invalid",
			idt:  auth.Identity{},
			err:  auth.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		idt, err := svc.Identify(context.Background(), tc.key)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.idt, idt, fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.idt, idt))
	}
}

func TestCreateGroup(t *testing.T) {
	svc := newService()
	_, secret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	key := auth.Key{
		ID:       "id",
		Type:     auth.APIKey,
		IssuerID: id,
		Subject:  email,
		IssuedAt: time.Now(),
	}

	_, apiToken, err := svc.Issue(context.Background(), secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing user's key expected to succeed: %s", err))

	parent := groups.Group{
		Name: "parent_group",
	}

	parentID, err := svc.CreateGroup(context.Background(), apiToken, parent)
	assert.Nil(t, err, fmt.Sprintf("Creating parent group failed: %s", err))

	group := groups.Group{
		Name: groupName,
	}

	cases := []struct {
		desc  string
		group groups.Group
		err   error
	}{
		{
			desc:  "create new group",
			group: group,
			err:   nil,
		},
		{
			desc:  "create group with existing name",
			group: group,
			err:   nil,
		},
		{
			desc: "create group with parent",
			group: groups.Group{
				Name:     groupName,
				ParentID: parentID,
			},
			err: nil,
		},
		{
			desc: "create group with invalid parent",
			group: groups.Group{
				Name:     groupName,
				ParentID: "xxxxxxxxxx",
			},
			err: groups.ErrCreateGroup,
		},
	}

	for _, tc := range cases {
		_, err := svc.CreateGroup(context.Background(), apiToken, tc.group)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

// func TestUpdateGroup(t *testing.T) {
// 	svc := newService()
// 	svc := newService(map[string]string{token: email})

// 	_, err := svc.Register(context.Background(), user)
// 	assert.Nil(t, err, fmt.Sprintf("registering user expected to succeed: %s", err))

// 	token, err := svc.Login(context.Background(), user)
// 	assert.Nil(t, err, fmt.Sprintf("authenticating user expected to succeed: %s", err))

// 	group := groups.Group{
// 		Name: groupName,
// 	}

// 	saved, err := svc.CreateGroup(context.Background(), token, group)
// 	assert.Nil(t, err, fmt.Sprintf("generating uuid expected to succeed: %s", err))

// 	group.Description = "test description"
// 	group.Name = "NewName"
// 	group.ID = saved.ID
// 	group.OwnerID = saved.OwnerID

// 	cases := []struct {
// 		desc  string
// 		group groups.Group
// 		err   error
// 	}{
// 		{
// 			desc:  "update group",
// 			group: group,
// 			err:   nil,
// 		},
// 	}

// 	for _, tc := range cases {
// 		err := svc.UpdateGroup(context.Background(), token, tc.group)
// 		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
// 		g, err := svc.ViewGroup(context.Background(), token, saved.ID)
// 		assert.Nil(t, err, fmt.Sprintf("retrieve group failed: %s", err))
// 		assert.Equal(t, tc.group.Description, g.Description, tc.desc, tc.err)
// 		assert.Equal(t, tc.group.Name, g.Name, tc.desc, tc.err)
// 		assert.Equal(t, tc.group.ID, g.ID, tc.desc, tc.err)
// 		assert.Equal(t, tc.group.OwnerID, g.OwnerID, tc.desc, tc.err)
// 	}
// }

// func TestRemoveGroup(t *testing.T) {
// 	svc := newService()

// 	_, err := svc.Register(context.Background(), user)
// 	assert.Nil(t, err, fmt.Sprintf("registering user expected to succeed: %s", err))

// 	token, err := svc.Login(context.Background(), user)
// 	assert.Nil(t, err, fmt.Sprintf("authenticating user expected to succeed: %s", err))

// 	group := groups.Group{
// 		Name: groupName,
// 	}

// 	saved, err := svc.CreateGroup(context.Background(), token, group)
// 	assert.Nil(t, err, fmt.Sprintf("generating uuid expected to succeed: %s", err))

// 	group.Description = "test description"
// 	group.Name = "NewName"
// 	group.ID = saved.ID
// 	group.OwnerID = saved.OwnerID

// 	cases := []struct {
// 		desc  string
// 		group groups.Group
// 		err   error
// 	}{
// 		{
// 			desc:  "remove existing group",
// 			group: group,
// 			err:   nil,
// 		},
// 		{
// 			desc:  "remove non existing group",
// 			group: group,
// 			err:   groups.ErrNotFound,
// 		},
// 	}

// 	for _, tc := range cases {
// 		err := svc.RemoveGroup(context.Background(), token, tc.group.ID)
// 		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
// 	}
// }
