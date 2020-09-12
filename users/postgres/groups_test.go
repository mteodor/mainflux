// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	uuidProvider "github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	groupName = "Mainflux"
	password  = "12345678"
)

func TestGroupSave(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("user id unexpected error: %s", err))
	user := users.User{
		ID:       uid,
		Email:    "TestGroupSave@mainflux.com",
		Password: password,
	}
	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	uid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group := users.Group{
		ID:    uid,
		Name:  "TestGroupSave",
		Owner: user,
	}

	cases := []struct {
		desc  string
		group users.Group
		err   error
	}{
		{
			desc:  "create new group",
			group: group,
			err:   nil,
		},
		{
			desc:  "create group that already exist",
			group: group,
			err:   users.ErrGroupConflict,
		},
		{
			desc: "create thing with invalid name",
			group: users.Group{
				Name: "x^%",
			},
			err: users.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		_, err := repo.Save(context.Background(), tc.group)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestGroupRetrieveByID(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user := users.User{
		ID:       uid,
		Email:    "TestGroupRetrieveByID@mainflux.com",
		Password: password,
	}
	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group1 := users.Group{
		ID:    gid,
		Name:  groupName + "TestGroupRetrieveByID1",
		Owner: user,
	}

	gid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group2 := users.Group{
		ID:    gid,
		Name:  groupName + "TestGroupRetrieveByID2",
		Owner: user,
	}

	g1, err := repo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	g2, err := repo.Save(context.Background(), group2)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	g2.ID, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("failed to generate id error: %s", err))

	cases := []struct {
		desc  string
		group users.Group
		err   error
	}{
		{
			desc:  "retrieve group for valid id",
			group: g1,
			err:   nil,
		},
		{
			desc:  "retrieve group for invalid id",
			group: g2,
			err:   users.ErrNotFound,
		},
	}

	for _, tc := range cases {
		_, err := repo.RetrieveByID(context.Background(), tc.group.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestGroupDelete(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user := users.User{
		ID:       uid,
		Email:    "TestGroupDelete@mainflux.com",
		Password: password,
	}
	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group1 := users.Group{
		ID:    gid,
		Name:  groupName + "TestGroupDelete1",
		Owner: user,
	}

	g1, err := repo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	err = repo.AssignUser(context.Background(), user.ID, g1.ID)
	require.Nil(t, err, fmt.Sprintf("failed to assign user to a group: %s", err))

	gid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group2 := users.Group{
		ID:    gid,
		Name:  groupName + "TestGroupDelete2",
		Owner: user,
	}

	g2, err := repo.Save(context.Background(), group2)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	cases := []struct {
		desc  string
		group users.Group
		err   error
	}{
		{
			desc:  "delete group for existing id",
			group: g2,
			err:   nil,
		},
		{
			desc:  "delete group for non-existing id",
			group: g2,
			err:   users.ErrDeleteGroupMissing,
		},
		{
			desc:  "delete group not empty",
			group: g1,
			err:   users.ErrDeleteGroupNotEmpty,
		},
	}

	for _, tc := range cases {
		err := repo.Delete(context.Background(), tc.group.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestAssignUser(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user := users.User{
		ID:       uid,
		Email:    "TestAssignUser@mainflux.com",
		Password: password,
	}

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group1 := users.Group{
		ID:    gid,
		Name:  groupName + "TestAssignUser1",
		Owner: user,
	}

	g1, err := repo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	gid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group2 := users.Group{
		ID:    gid,
		Name:  groupName + "TestAssignUser2",
		Owner: user,
	}

	g2, err := repo.Save(context.Background(), group2)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	gid, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id generating error: %s", err))
	g3 := users.Group{
		ID: gid,
	}

	cases := []struct {
		desc  string
		group users.Group
		err   error
	}{
		{
			desc:  "assign user to existing group",
			group: g1,
			err:   nil,
		},
		{
			desc:  "assign user to another existing group",
			group: g2,
			err:   nil,
		},
		{
			desc:  "assign already assigned user to a group",
			group: g2,
			err:   users.ErrUserAlreadyAssigned,
		},
		{
			desc:  "assign user to non existing group",
			group: g3,
			err:   users.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := repo.AssignUser(context.Background(), user.ID, tc.group.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}
