// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	uuidProvider "github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	metadata  = users.Metadata{"field": "value"}
	wrongMeta = users.Metadata{"wrong": "wrong"}
	user      = users.User{
		Email:    "example@mainflux.com",
		Password: password,
	}
	nonExistingUser = users.User{
		Email:    "non_existing@mainflux.com",
		Password: password,
	}
)

const (
	groupName              = "Mainflux"
	password               = "12345678"
	metaNum                = 5
	numOfGroups            = 10
	numOfAncestorsInSubset = 5
)

func TestGroupSave(t *testing.T) {
	err := cleanDB(db)
	require.Nil(t, err, fmt.Sprintf("error cleaning db: %s", err))
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("user id unexpected error: %s", err))
	user.ID = uid

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	uid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group := users.Group{
		ID:      uid,
		Name:    "TestGroupSave",
		OwnerID: user.ID,
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
	err := cleanDB(db)
	require.Nil(t, err, fmt.Sprintf("error cleaning db: %s", err))
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user.ID = uid

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group1 := users.Group{
		ID:      gid,
		Name:    groupName + "TestGroupRetrieveByID1",
		OwnerID: user.ID,
	}

	gid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group2 := users.Group{
		ID:      gid,
		Name:    groupName + "TestGroupRetrieveByID2",
		OwnerID: user.ID,
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
	err := cleanDB(db)
	require.Nil(t, err, fmt.Sprintf("error cleaning db: %s", err))
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user.ID = uid

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group1 := users.Group{
		ID:      gid,
		Name:    groupName + "TestGroupDelete1",
		OwnerID: user.ID,
	}

	g1, err := repo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	err = repo.Assign(context.Background(), user.ID, g1.ID)
	require.Nil(t, err, fmt.Sprintf("failed to assign user to a group: %s", err))

	gid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group2 := users.Group{
		ID:      gid,
		Name:    groupName + "TestGroupDelete2",
		OwnerID: user.ID,
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
	}

	for _, tc := range cases {
		err := repo.Delete(context.Background(), tc.group.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestAssignUser(t *testing.T) {
	err := cleanDB(db)
	require.Nil(t, err, fmt.Sprintf("error cleaning db: %s", err))
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user.ID = uid

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group1 := users.Group{
		ID:      gid,
		Name:    groupName + "TestAssignUser1",
		OwnerID: user.ID,
	}

	g1, err := repo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	gid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group2 := users.Group{
		ID:      gid,
		Name:    groupName + "TestAssignUser2",
		OwnerID: user.ID,
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
			desc:  "assign user to non existing group",
			group: g3,
			err:   users.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := repo.Assign(context.Background(), user.ID, tc.group.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}

func TestUnassignUser(t *testing.T) {
	err := cleanDB(db)
	require.Nil(t, err, fmt.Sprintf("error cleaning db: %s", err))
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)

	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user.ID = uid

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user1, err := userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	uid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	nonExistingUser.ID = uid

	_, err = userRepo.Save(context.Background(), nonExistingUser)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	nonExistingUser, err := userRepo.RetrieveByEmail(context.Background(), nonExistingUser.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group := users.Group{
		ID:      gid,
		Name:    groupName,
		OwnerID: user.ID,
	}

	g1, err := repo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	err = repo.Assign(context.Background(), user1.ID, group.ID)
	require.Nil(t, err, fmt.Sprintf("failed to assign user: %s", err))

	cases := []struct {
		desc  string
		group users.Group
		user  users.User
		err   error
	}{
		{desc: "remove user from a group", group: g1, user: user1, err: nil},
		{desc: "remove already removed user from a group", group: g1, user: user1, err: nil},
		{desc: "remove non existing user from a group", group: g1, user: nonExistingUser, err: nil},
	}

	for _, tc := range cases {
		err := repo.Unassign(context.Background(), tc.user.ID, tc.group.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}

func TestRetrieveAll(t *testing.T) {
	err := cleanDB(db)
	require.Nil(t, err, fmt.Sprintf("error cleaning db: %s", err))
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user.ID = uid

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	n := uint64(numOfGroups)
	for i := uint64(0); i < n; i++ {
		gid, err := uuid.New().ID()
		require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
		group := users.Group{
			ID:      gid,
			Name:    fmt.Sprintf("group%d", i),
			OwnerID: user.ID,
		}

		// Create Groups with metadata.
		if i < metaNum {
			group.Metadata = metadata
		}

		_, err = repo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
	}

	cases := map[string]struct {
		owner    string
		offset   uint64
		limit    uint64
		name     string
		size     uint64
		total    uint64
		metadata map[string]interface{}
	}{
		"retrieve all groups": {
			offset: 0,
			limit:  n,
			size:   n,
			total:  n,
		},
		"retrieve subset of groups": {
			offset: n / 2,
			limit:  n,
			size:   n / 2,
			total:  n,
		},
		"retrieve groups with existing metadata": {
			offset:   0,
			limit:    n,
			size:     metaNum,
			total:    metaNum,
			metadata: metadata,
		},
		"retrieve groups with non-existing metadata": {
			offset:   0,
			limit:    n,
			size:     0,
			total:    0,
			metadata: wrongMeta,
		},
	}
	for desc, tc := range cases {
		page, err := repo.RetrieveAll(context.Background(), tc.offset, tc.limit, tc.metadata)
		size := uint64(len(page.Groups))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.size, size))
		assert.Equal(t, tc.total, page.Total, fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.total, page.Total))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}

}

func TestRetrieveAllAncestors(t *testing.T) {
	err := cleanDB(db)
	require.Nil(t, err, fmt.Sprintf("error cleaning db: %s", err))
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user.ID = uid

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	var parentID = ""
	var lastChildParentID = ""
	var subsetAncestorID = ""
	n := uint64(numOfGroups)
	for i := uint64(1); i <= n; i++ {
		gid, err := uuid.New().ID()
		require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
		group := users.Group{
			ID:       gid,
			Name:     fmt.Sprintf("group%d", i),
			OwnerID:  user.ID,
			ParentID: parentID,
		}

		// Create Groups with metadata.
		if i <= metaNum {
			group.Metadata = metadata
		}

		g, err := repo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
		if i == numOfGroups {
			lastChildParentID = g.ParentID
		}
		if i == numOfAncestorsInSubset {
			subsetAncestorID = g.ID
		}
		parentID = g.ID
	}
	cases := map[string]struct {
		parentID string
		offset   uint64
		limit    uint64
		name     string
		size     uint64
		total    uint64
		metadata map[string]interface{}
	}{
		"retrieve all groups": {
			parentID: lastChildParentID,
			offset:   0,
			limit:    n,
			size:     n,
			total:    n,
		},
		"retrieve subset of groups": {
			parentID: lastChildParentID,
			offset:   n / 2,
			limit:    n,
			size:     n / 2,
			total:    n,
		},
		"retrieve groups with existing metadata": {
			parentID: lastChildParentID,
			offset:   0,
			limit:    n,
			size:     metaNum,
			total:    metaNum,
			metadata: metadata,
		},
		"retrieve groups with non-existing metadata": {
			parentID: lastChildParentID,
			offset:   0,
			limit:    n,
			size:     0,
			total:    0,
			metadata: wrongMeta,
		},
		"retrieve subset of groups by different parent": {
			parentID: subsetAncestorID,
			offset:   0,
			limit:    n,
			size:     numOfAncestorsInSubset + 1,
			total:    numOfAncestorsInSubset + 1,
			metadata: nil,
		},
	}
	for desc, tc := range cases {
		page, err := repo.RetrieveAllAncestors(context.Background(), tc.parentID, tc.offset, tc.limit, tc.metadata)
		size := uint64(len(page.Groups))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.size, size))
		assert.Equal(t, tc.total, page.Total, fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.total, page.Total))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}

}

func TestRetrieveAllChildren(t *testing.T) {
	err := cleanDB(db)
	require.Nil(t, err, fmt.Sprintf("error cleaning db: %s", err))
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user.ID = uid

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user, err = userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	var parentID = ""
	var rootParentID = ""
	var subsetAncestorID = ""
	n := uint64(numOfGroups)
	for i := uint64(1); i <= n; i++ {
		gid, err := uuid.New().ID()
		require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
		group := users.Group{
			ID:       gid,
			Name:     fmt.Sprintf("group%d", i),
			OwnerID:  user.ID,
			ParentID: parentID,
		}

		// Create Groups with metadata.
		if i <= metaNum {
			group.Metadata = metadata
		}

		g, err := repo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
		if i == 1 {
			rootParentID = g.ID
		}
		if i == numOfAncestorsInSubset {
			subsetAncestorID = g.ID
		}
		parentID = g.ID
	}
	cases := map[string]struct {
		parentID string
		offset   uint64
		limit    uint64
		name     string
		size     uint64
		total    uint64
		metadata map[string]interface{}
	}{
		"retrieve all groups": {
			parentID: rootParentID,
			offset:   0,
			limit:    n,
			size:     n,
			total:    n,
		},
		"retrieve subset of groups": {
			parentID: rootParentID,
			offset:   n / 2,
			limit:    n,
			size:     n / 2,
			total:    n,
		},
		"retrieve groups with existing metadata": {
			parentID: rootParentID,
			offset:   0,
			limit:    n,
			size:     metaNum,
			total:    metaNum,
			metadata: metadata,
		},
		"retrieve groups with non-existing metadata": {
			parentID: rootParentID,
			offset:   0,
			limit:    n,
			size:     0,
			total:    0,
			metadata: wrongMeta,
		},
		"retrieve subset of groups by different parent": {
			parentID: subsetAncestorID,
			offset:   0,
			limit:    n,
			size:     numOfAncestorsInSubset + 1,
			total:    numOfAncestorsInSubset + 1,
			metadata: nil,
		},
	}
	for desc, tc := range cases {
		page, err := repo.RetrieveAllChildren(context.Background(), tc.parentID, tc.offset, tc.limit, tc.metadata)
		size := uint64(len(page.Groups))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.size, size))
		assert.Equal(t, tc.total, page.Total, fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.total, page.Total))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}

}

func cleanDB(db *sqlx.DB) error {
	if _, err := db.Exec("delete from groups"); err != nil {
		return err
	}
	if _, err := db.Exec("delete from users"); err != nil {
		return err
	}
	return nil
}
