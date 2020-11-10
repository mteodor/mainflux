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
	"github.com/mainflux/mainflux/things/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxLevel = 5

var (
	wrongMeta = groups.Metadata{"department": "wrong"}
	metadata  = groups.Metadata{"department": "IoT"}
)

func TestGroupsSave(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)
	cleanAll(t, dbMiddleware)

	ownerID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupID, err := uuidProvider.New().ID()
	fmt.Println(groupID)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupNoParent := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux",
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
		Name:        "mainflux_parent",
	}

	pg, err := groupRepo.Save(context.Background(), parentGroup)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	groupID, err = uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	group := groups.Group{
		Description: "description",
		ID:          groupID,
		OwnerID:     ownerID,
		Name:        "mainflux_with_parent",
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
		Name:        "mainflux_invalid",
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
			err:   groups.ErrMissingParent,
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
	cleanAll(t, dbMiddleware)

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
			err:     groups.ErrNotFound,
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
	cleanAll(t, dbMiddleware)
	ownerID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	up := uuidProvider.New()
	nameNum := uint64(3)
	metaNum := uint64(3)

	n := uint64(10)
	for i := uint64(0); i < n; i++ {
		groupID, err := up.ID()
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		group := groups.Group{
			Description: "description",
			ID:          groupID,
			OwnerID:     ownerID,
			Name:        fmt.Sprintf("mainflux%d", i),
		}

		// Create Things with metadata.
		if i >= nameNum && i < nameNum+metaNum {
			group.Metadata = metadata
		}

		_, err = groupRepo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
	}

	// Create group with parents to test level
	parentID := ""
	for i := n; i < n+maxLevel; i++ {
		groupID, err := up.ID()
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		group := groups.Group{
			Description: "description",
			ID:          groupID,
			OwnerID:     ownerID,
			Name:        fmt.Sprintf("mainflux%d", i),
			ParentID:    parentID,
		}

		// Create Things with metadata.
		if i >= nameNum && i < nameNum+metaNum {
			group.Metadata = metadata
		}

		_, err = groupRepo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
		parentID = groupID
	}

	cases := map[string]struct {
		name     string
		level    uint64
		total    uint64
		metadata map[string]interface{}
	}{
		"retrieve all groups": {
			level: maxLevel,
			total: n + maxLevel,
		},
		"retrieve subset of groups": {
			level: maxLevel - 1,
			total: n + maxLevel - 1,
		},
		"retrieve groups with existing metadata": {
			level:    maxLevel,
			total:    metaNum,
			metadata: metadata,
		},
		"retrieve things with non-existing metadata": {
			level:    maxLevel,
			total:    0,
			metadata: wrongMeta,
		},
	}

	for desc, tc := range cases {
		page, err := groupRepo.RetrieveAll(context.Background(), tc.level, tc.metadata)
		assert.Equal(t, tc.total, countTotal(page.Groups), fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.total, countTotal(page.Groups)))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestRetrieveAllParents(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)
	cleanAll(t, dbMiddleware)
	ownerID, err := uuidProvider.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	up := uuidProvider.New()
	metaNum := uint64(5)

	// Create groups with parent
	parentID := ""
	n := uint64(10)
	path := ""
	for i := uint64(0); i < n; i++ {
		groupID, err := up.ID()
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		group := groups.Group{
			Description: "description",
			ID:          groupID,
			OwnerID:     ownerID,
			Name:        fmt.Sprintf("mainflux%d", i),
			ParentID:    parentID,
		}
		path = path + group.Name + "."
		// Create Things with metadata.
		if i >= metaNum {
			group.Metadata = metadata
		}

		_, err = groupRepo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
		parentID = groupID
	}

	cases := map[string]struct {
		name     string
		level    uint64
		total    uint64
		parentID string
		metadata map[string]interface{}
	}{
		"retrieve all parents up to max level including itself": {
			level:    maxLevel,
			parentID: parentID,
			total:    maxLevel + 1,
		},
		"retrieve parents up to level including itself": {
			level:    maxLevel - 1,
			parentID: parentID,
			total:    maxLevel,
		},
		"retrieve parents with existing metadata": {
			level:    maxLevel,
			total:    metaNum,
			parentID: parentID,
			metadata: metadata,
		},
		"retrieve things with non-existing metadata": {
			level:    maxLevel,
			total:    0,
			parentID: parentID,
			metadata: wrongMeta,
		},
	}

	for desc, tc := range cases {
		page, err := groupRepo.RetrieveAllParents(context.Background(), tc.parentID, tc.level, tc.metadata)
		assert.Equal(t, tc.total, countTotal(page.Groups), fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.total, countTotal(page.Groups)))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestGroupDelete(t *testing.T) {
	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.NewGroupRepo(dbMiddleware)
	cleanAll(t, dbMiddleware)
	gid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	ownerID, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	group := groups.Group{
		Description: "description",
		ID:          gid,
		OwnerID:     ownerID,
		Name:        "mainflux_no_parent",
	}
	_, err = repo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("save got unexpected error: %s", err))

	gidNonExisting, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc    string
		groupID string
		err     error
	}{
		{
			desc:    "delete group for existing id",
			groupID: gid,
			err:     nil,
		},
		{
			desc:    "delete group for non-existing id",
			groupID: gidNonExisting,
			err:     groups.ErrDeleteGroup,
		},
	}

	for _, tc := range cases {
		err := repo.Delete(context.Background(), tc.groupID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func cleanAll(t *testing.T, db postgres.Database) {
	params := map[string]interface{}{}
	rows, err := db.NamedQueryContext(context.Background(), "delete from thing_groups;", params)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	defer rows.Close()
	total := uint64(0)
	if rows.Next() {
		err := rows.Scan(&total)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	}
}

func countTotal(groups []groups.Group) uint64 {
	size := uint64(0)
	for _, g := range groups {
		size = size + countChildrenSize(g.Children)
	}
	return size + uint64(len(groups))
}

func countChildrenSize(groups []*groups.Group) uint64 {
	size := uint64(0)
	if len(groups) == 0 {
		return 0
	}
	for _, g := range groups {
		size = size + countChildrenSize(g.Children)
	}
	return size + uint64(len(groups))
}
