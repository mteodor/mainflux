// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/internal/groups"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	uuidProvider "github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/things/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
	// Save group
	Save(ctx context.Context, group Group) (Group, error)

	// Update a group
	Update(ctx context.Context, group Group) (Group, error)

	// Delete a group
	Delete(ctx context.Context, groupID string) error

	// RetrieveByID retrieves group by its id
	RetrieveByID(ctx context.Context, id string) (Group, error)

	// RetrieveByName retrieves group by its name
	RetrieveByName(ctx context.Context, name string) (Group, error)

	// RetrieveAll retrieves all groups.
	RetrieveAll(ctx context.Context, offset, limit uint64, gm Metadata) (GroupPage, error)

	// RetrieveAllParents retrieves all groups that are ancestors to the group with given groupID.
	RetrieveAllParents(ctx context.Context, groupID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// RetrieveAllChildren retrieves all children from group with given groupID.
	RetrieveAllChildren(ctx context.Context, groupID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// Retrieves list of groups that member belongs to
	Memberships(ctx context.Context, memberID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// Members retrieves everything that is assigned to a group identified by groupID.
	Members(ctx context.Context, groupID string, offset, limit uint64, meta Metadata) (Page, error)

	// Assign adds member to group.
	Assign(ctx context.Context, memberID, groupID string) error

	// Unassign removes a member from a group
	Unassign(ctx context.Context, memberID, groupID string) error
*/

var (
	metadata = things.Metadata{
		"field": "value",
	}
	wrongMeta = things.Metadata{
		"wrong": "wrong",
	}
)

func TestGroupsSave(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	ownerID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupID, err := uuidProvider.New().ID()
	fmt.Println(groupID)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupNoParent := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux-no-parent",
	}

	groupID, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupNoParentSameName := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux",
	}

	groupID, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	parentGroup := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux-parent",
	}

	pg, err := groupRepo.Save(context.Background(), parentGroup)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupID, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	group := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux-with-parent",
		ParentID:    pg.ID,
	}

	groupID, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupNoParentSameName = groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux",
	}

	groupID, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	invalidParentID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupInvalidParent := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux-invalid",
		ParentID:    invalidParentID,
	}

	cases := []struct {
		desc  string
		group groups.Group
		err   error
	}{
		{
			desc:  "create new group",
			group: groupNoParent,
			err:   nil,
		},
		{
			desc:  "create group that already exist",
			group: groupNoParentSameName,
			err:   groups.ErrGroupConflict,
		},

		{
			desc:  "create group with parent id",
			group: group,
			err:   nil,
		},
		{
			desc:  "create group with invalid parentID",
			group: groupInvalidParent,
			err:   groups.ErrCreateGroup,
		},
	}

	for _, tc := range cases {
		_, err := groupRepo.Save(context.Background(), tc.group)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("\"%s\": expected \"%s\" got \"%s\"\n", tc.desc, tc.err, err))
	}
}

func TestRetrieveByID(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	ownerID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	group := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux",
	}

	_, err = groupRepo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	incorrectID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc    string
		groupID string
		err     error
	}{
		{
			desc:    "retrieve a group",
			groupID: groupID,
			err:     nil,
		},
		{
			desc:    "retrieve a group with incorrect id",
			groupID: incorrectID,
			err:     things.ErrNotFound,
		},
	}

	for _, tc := range cases {
		_, err := groupRepo.RetrieveByID(context.Background(), tc.groupID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("\"%s\": expected \"%s\" got \"%s\"\n", tc.desc, tc.err, err))
	}
}

func TestRetrieveAll(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	ownerID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	group := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux-no-parent",
	}

	_, err = groupRepo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	incorrectID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	up := uuidProvider.New()
	offset := uint64(1)
	nameNum := uint64(3)
	metaNum := uint64(3)
	nameMetaNum := uint64(2)

	n := uint64(10)
	for i := uint64(0); i < n; i++ {
		id, err := up.ID()
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		key, err := up.ID()
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		th := things.Thing{
			Owner: email,
			ID:    id,
			Key:   key,
		}

		// Create Things with name.
		if i < nameNum {
			th.Name = name
		}
		// Create Things with metadata.
		if i >= nameNum && i < nameNum+metaNum {
			th.Metadata = metadata
		}
		// Create Things with name and metadata.
		if i >= n-nameMetaNum {
			th.Metadata = metadata
			th.Name = name
		}

		_, err = thingRepo.Save(context.Background(), th)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
	}

	cases := map[string]struct {
		offset   uint64
		limit    uint64
		name     string
		size     uint64
		total    uint64
		metadata map[string]interface{}
	}{
		"retrieve all things with existing owner": {
			offset: 0,
			limit:  n,
			size:   n,
			total:  n,
		},
		"retrieve subset of things with existing owner": {
			offset: n / 2,
			limit:  n,
			size:   n / 2,
			total:  n,
		},
		"retrieve things with non-existing owner": {
			offset: 0,
			limit:  n,
			size:   0,
			total:  0,
		},
		"retrieve things with existing name": {
			offset: 1,
			limit:  n,
			name:   name,
			size:   nameNum + nameMetaNum - offset,
			total:  nameNum + nameMetaNum,
		},
		"retrieve things with non-existing name": {
			offset: 0,
			limit:  n,
			name:   "wrong",
			size:   0,
			total:  0,
		},
		"retrieve things with existing metadata": {
			offset:   0,
			limit:    n,
			size:     metaNum + nameMetaNum,
			total:    metaNum + nameMetaNum,
			metadata: metadata,
		},
		"retrieve things with non-existing metadata": {
			offset:   0,
			limit:    n,
			size:     0,
			total:    0,
			metadata: wrongMeta,
		},
		"retrieve all things with existing name and metadata": {
			offset:   0,
			limit:    n,
			size:     nameMetaNum,
			total:    nameMetaNum,
			name:     name,
			metadata: metadata,
		},
	}

	for desc, tc := range cases {
		page, err := groupRepo.RetrieveAll(context.Background(), tc.offset, tc.limit, tc.metadata)
		size := uint64(len(page.Groups))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.size, size))
		assert.Equal(t, tc.total, page.Total, fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.total, page.Total))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestGroupDelete(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user := things.User{
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
	group1 := groups.Group{
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
	group2 := groups.Group{
		ID:      gid,
		Name:    groupName + "TestGroupDelete2",
		OwnerID: user.ID,
	}

	g2, err := repo.Save(context.Background(), group2)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	cases := []struct {
		desc  string
		group groups.Group
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
			err:   things.ErrDeleteGroupMissing,
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
	user := things.User{
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
	group1 := groups.Group{
		ID:      gid,
		Name:    groupName + "TestAssignUser1",
		OwnerID: user.ID,
	}

	g1, err := repo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	gid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group2 := groups.Group{
		ID:      gid,
		Name:    groupName + "TestAssignUser2",
		OwnerID: user.ID,
	}

	g2, err := repo.Save(context.Background(), group2)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	gid, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id generating error: %s", err))
	g3 := groups.Group{
		ID: gid,
	}

	cases := []struct {
		desc  string
		group groups.Group
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
			err:   things.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := repo.Assign(context.Background(), user.ID, tc.group.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}

func TestUnassignUser(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	userRepo := postgres.NewUserRepo(dbMiddleware)

	uid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user := things.User{
		ID:       uid,
		Email:    "UnassignUser1@mainflux.com",
		Password: password,
	}

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user1, err := userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	uid, err = uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	user = things.User{
		ID:       uid,
		Email:    "UnassignUser2@mainflux.com",
		Password: password,
	}

	_, err = userRepo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	user2, err := userRepo.RetrieveByEmail(context.Background(), user.Email)
	require.Nil(t, err, fmt.Sprintf("retrieve got unexpected error: %s", err))

	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group1 := groups.Group{
		ID:      gid,
		Name:    groupName + "UnassignUser1",
		OwnerID: user.ID,
	}

	g1, err := repo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	err = repo.Assign(context.Background(), user1.ID, group1.ID)
	require.Nil(t, err, fmt.Sprintf("failed to assign user: %s", err))

	cases := []struct {
		desc  string
		group groups.Group
		user  things.User
		err   error
	}{
		{desc: "remove user from a group", group: g1, user: user1, err: nil},
		{desc: "remove already removed user from a group", group: g1, user: user1, err: nil},
		{desc: "remove non existing user from a group", group: g1, user: user2, err: nil},
	}

	for _, tc := range cases {
		err := repo.Unassign(context.Background(), tc.user.ID, tc.group.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}
