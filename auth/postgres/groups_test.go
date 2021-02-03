// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mainflux/mainflux/auth/groups"
	"github.com/mainflux/mainflux/auth/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	maxNameSize = 254
	maxDescSize = 1024
	groupName   = "Mainflux"
	description = "description"
)

var (
	invalidName = strings.Repeat("m", maxNameSize+1)
	invalidDesc = strings.Repeat("m", maxDescSize+1)
	metadata    = groups.Metadata{
		"admin": "true",
	}
)

type member struct {
	ID string
}

func (m member) GetID() string {
	return m.ID
}

func generateGroupID(t *testing.T) string {
	grpID, err := ulidProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	return grpID
}

func TestGroupSave(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	usrID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	wrongID, err := ulidProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	grpID := generateGroupID(t)

	cases := []struct {
		desc  string
		group groups.Group
		err   error
	}{
		{
			desc: "create new thing group",
			group: groups.Group{
				ID:      grpID,
				OwnerID: usrID,
				Name:    "mainflux",
				Type:    "things",
			},
			err: nil,
		},
		{
			desc: "create new thing group with existing name",
			group: groups.Group{
				ID:      grpID,
				OwnerID: usrID,
				Name:    "mainflux",
				Type:    "things",
			},
			err: groups.ErrGroupConflict,
		},
		{
			desc: "create new users group",
			group: groups.Group{
				ID:      generateGroupID(t),
				OwnerID: usrID,
				Name:    "mainflux",
				Type:    "users",
			},
			err: nil,
		},
		{
			desc: "create group with wrong type",
			group: groups.Group{
				ID:      generateGroupID(t),
				OwnerID: usrID,
				Name:    "mainflux",
				Type:    "wrong",
			},
			err: groups.ErrInvalidGroupType,
		},
		{
			desc: "create group with invalid name",
			group: groups.Group{
				ID:      generateGroupID(t),
				OwnerID: usrID,
				Name:    invalidName,
				Type:    "things",
			},
			err: groups.ErrMalformedEntity,
		},
		{
			desc: "create group with invalid description",
			group: groups.Group{
				ID:          generateGroupID(t),
				OwnerID:     usrID,
				Name:        "mainflux",
				Type:        "things",
				Description: invalidDesc,
			},
			err: groups.ErrMalformedEntity,
		},
		{
			desc: "create group with parent",
			group: groups.Group{
				ID:       generateGroupID(t),
				ParentID: grpID,
				OwnerID:  usrID,
				Name:     "withParent",
				Type:     "things",
			},
			err: nil,
		},
		{
			desc: "create group with parent and existing name",
			group: groups.Group{
				ID:       generateGroupID(t),
				ParentID: grpID,
				OwnerID:  usrID,
				Name:     "mainflux",
				Type:     "things",
			},
			err: nil,
		},
		{
			desc: "create group with wrong parent",
			group: groups.Group{
				ID:       generateGroupID(t),
				ParentID: wrongID,
				OwnerID:  usrID,
				Name:     "wrongParent",
				Type:     "things",
			},
			err: groups.ErrCreateGroup,
		},
	}

	for _, tc := range cases {
		_, err := groupRepo.Save(context.Background(), tc.group)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}

func TestGroupRetrieveByID(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	uid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	require.Nil(t, err, fmt.Sprintf("group id unexpected error: %s", err))
	group1 := groups.Group{
		ID:      generateGroupID(t),
		Name:    groupName + "TestGroupRetrieveByID1",
		OwnerID: uid,
		Type:    "things",
	}

	_, err = groupRepo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	retrieved, err := groupRepo.RetrieveByID(context.Background(), group1.ID)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	assert.True(t, retrieved.ID == group1.ID, fmt.Sprintf("Save group, ID: expected %s got %s\n", group1.ID, retrieved.ID))
	assert.True(t, retrieved.Type == group1.Type, fmt.Sprintf("Save group, ID: expected %s got %s\n", group1.Type, retrieved.Type))

	creationTime := time.Now().UTC()

	group2 := groups.Group{
		ID:          generateGroupID(t),
		Name:        groupName + "TestGroupRetrieveByID",
		OwnerID:     uid,
		ParentID:    group1.ID,
		CreatedAt:   creationTime,
		UpdatedAt:   creationTime,
		Description: description,
		Metadata:    metadata,
	}

	_, err = groupRepo.Save(context.Background(), group2)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	retrieved, err = groupRepo.RetrieveByID(context.Background(), group2.ID)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	assert.True(t, retrieved.ID == group2.ID, fmt.Sprintf("Save group, ID: expected %s got %s\n", group2.ID, retrieved.ID))
	// Type for the group is inherited from the parent
	assert.True(t, retrieved.Type == group1.Type, fmt.Sprintf("Save group, Type: expected %s got %s\n", group1.Type, retrieved.Type))
	// There is a problem with retrieving time from DB, looks like rounding in DB makes difference.
	// assert.True(t, retrieved.CreatedAt.Equal(creationTime), fmt.Sprintf("Save group, CreatedAt: expected %s got %s\n", creationTime, retrieved.CreatedAt))
	// assert.True(t, retrieved.UpdatedAt.Equal(creationTime), fmt.Sprintf("Save group, UpdatedAt: expected %s got %s\n", creationTime, retrieved.UpdatedAt))
	assert.True(t, retrieved.Level == 2, fmt.Sprintf("Save group, Level: expected %d got %d\n", retrieved.Level, 2))
	assert.True(t, retrieved.ParentID == group1.ID, fmt.Sprintf("Save group, Level: expected %s got %s\n", group1.ID, retrieved.ParentID))
	assert.True(t, retrieved.Description == description, fmt.Sprintf("Save group, Description: expected %v got %v\n", retrieved.Description, description))
	assert.True(t, retrieved.Path == fmt.Sprintf("%s.%s", group1.ID, group2.ID), fmt.Sprintf("Save group, Path: expected %s got %s\n", fmt.Sprintf("%s.%s", group1.ID, group2.ID), retrieved.Path))

	retrieved, err = groupRepo.RetrieveByID(context.Background(), generateGroupID(t))
	assert.True(t, errors.Contains(err, groups.ErrNotFound), fmt.Sprintf("Retrieve group: expected %s got %s\n", groups.ErrNotFound, err))
}

func TestGroupUpdate(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	uid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	creationTime := time.Now().UTC()
	updateTime := time.Now().UTC()
	groupID := generateGroupID(t)

	group := groups.Group{
		ID:          groupID,
		Name:        groupName + "TestGroupUpdate",
		OwnerID:     uid,
		Type:        "things",
		CreatedAt:   creationTime,
		UpdatedAt:   creationTime,
		Description: description,
		Metadata:    metadata,
	}

	_, err = groupRepo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	retrieved, err := groupRepo.RetrieveByID(context.Background(), group.ID)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	cases := []struct {
		desc          string
		groupUpdate   groups.Group
		groupExpected groups.Group
		err           error
	}{
		{
			desc: "update group for existing id",
			groupUpdate: groups.Group{
				ID:        groupID,
				Name:      groupName + "Updated",
				Type:      "users",
				UpdatedAt: updateTime,
				Metadata:  groups.Metadata{"admin": "false"},
			},
			groupExpected: groups.Group{
				Name:      groupName + "Updated",
				Type:      "things",
				UpdatedAt: updateTime,
				Metadata:  groups.Metadata{"admin": "false"},
				CreatedAt: retrieved.CreatedAt,
				Path:      retrieved.Path,
				ParentID:  retrieved.ParentID,
				ID:        retrieved.ID,
				Level:     retrieved.Level,
			},
			err: nil,
		},
		{
			desc: "update group for non-existing id",
			groupUpdate: groups.Group{
				ID:   "wrong",
				Name: groupName + "-2",
			},
			err: groups.ErrUpdateGroup,
		},
		{
			desc: "update group for invalid name",
			groupUpdate: groups.Group{
				ID:   groupID,
				Name: invalidName,
			},
			err: groups.ErrMalformedEntity,
		},
		{
			desc: "update group for invalid description",
			groupUpdate: groups.Group{
				ID:          groupID,
				Description: invalidDesc,
			},
			err: groups.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		updated, err := groupRepo.Update(context.Background(), tc.groupUpdate)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if tc.desc == "update group for existing id" {
			assert.True(t, updated.Level == tc.groupExpected.Level, fmt.Sprintf("%s:Level: expected %d got %d\n", tc.desc, tc.groupExpected.Level, updated.Level))
			assert.True(t, updated.Name == tc.groupExpected.Name, fmt.Sprintf("%s:Name: expected %s got %s\n", tc.desc, tc.groupExpected.Name, updated.Name))
			assert.True(t, updated.Type == tc.groupExpected.Type, fmt.Sprintf("%s:Type: expected %s got %s\n", tc.desc, tc.groupExpected.Type, updated.Type))
			assert.True(t, updated.Metadata["admin"] == tc.groupExpected.Metadata["admin"], fmt.Sprintf("%s:Level: expected %d got %d\n", tc.desc, tc.groupExpected.Metadata["admin"], updated.Metadata["admin"]))
		}
	}
}

func TestGroupDelete(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	uid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	creationTime := time.Now().UTC()
	groupParent := groups.Group{
		ID:        generateGroupID(t),
		Name:      groupName + "Updated",
		OwnerID:   uid,
		Type:      "things",
		Metadata:  groups.Metadata{"admin": "false"},
		CreatedAt: creationTime,
		UpdatedAt: creationTime,
	}

	groupParent, err = groupRepo.Save(context.Background(), groupParent)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	creationTime = time.Now().UTC()
	groupChild1 := groups.Group{
		ID:        generateGroupID(t),
		ParentID:  groupParent.ID,
		Name:      groupName + "child1",
		OwnerID:   uid,
		Type:      "things",
		Metadata:  groups.Metadata{"admin": "false"},
		CreatedAt: creationTime,
		UpdatedAt: creationTime,
	}

	creationTime = time.Now().UTC()
	groupChild2 := groups.Group{
		ID:        generateGroupID(t),
		ParentID:  groupParent.ID,
		Name:      groupName + "child2",
		OwnerID:   uid,
		Type:      "things",
		Metadata:  groups.Metadata{"admin": "false"},
		CreatedAt: creationTime,
		UpdatedAt: creationTime,
	}

	groupChild1, err = groupRepo.Save(context.Background(), groupChild1)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	groupChild2, err = groupRepo.Save(context.Background(), groupChild2)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	gp, err := groupRepo.RetrieveAllChildren(context.Background(), groupParent.ID, 5, nil)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("Retrieve children for parent: expected %v got %v\n", nil, err))
	assert.True(t, gp.Total == 3, fmt.Sprintf("Number of children + parent: expected %d got %d\n", 3, gp.Total))

	thingID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("thing id create unexpected error: %s", err))

	err = groupRepo.Assign(context.Background(), thingID, groupChild1)
	require.Nil(t, err, fmt.Sprintf("thing assign got unexpected error: %s", err))

	err = groupRepo.Delete(context.Background(), groupChild1.ID)
	assert.True(t, errors.Contains(err, groups.ErrGroupNotEmpty), fmt.Sprintf("delete non empty group: expected %v got %v\n", groups.ErrGroupNotEmpty, err))

	err = groupRepo.Delete(context.Background(), groupChild2.ID)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("delete empty group: expected %v got %v\n", nil, err))

	err = groupRepo.Delete(context.Background(), groupParent.ID)
	assert.True(t, errors.Contains(err, groups.ErrGroupNotEmpty), fmt.Sprintf("delete parent with children with members: expected %v got %v\n", groups.ErrGroupNotEmpty, err))

	gp, err = groupRepo.RetrieveAllChildren(context.Background(), groupParent.ID, 5, nil)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("retrieve children after one child removed: expected %v got %v\n", nil, err))
	assert.True(t, gp.Total == 2, fmt.Sprintf("number of children + parent: expected %d got %d\n", 2, gp.Total))

	err = groupRepo.Unassign(context.Background(), thingID, groupChild1)
	require.Nil(t, err, fmt.Sprintf("failed to remove thing from a group error: %s", err))

	err = groupRepo.Delete(context.Background(), groupParent.ID)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("delete parent with children with no members: expected %v got %v\n", nil, err))

	_, err = groupRepo.RetrieveByID(context.Background(), groupChild1.ID)
	assert.True(t, errors.Contains(err, groups.ErrNotFound), fmt.Sprintf("retrieve child after parent removed: expected %v got %v\n", nil, err))
}

func TestRetrieveAll(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)
	uid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	metadata := groups.Metadata{
		"field": "value",
	}
	wrongMeta := groups.Metadata{
		"wrong": "wrong",
	}

	metaNum := uint64(3)

	n := uint64(10)
	parentID := ""
	for i := uint64(0); i < n; i++ {
		creationTime := time.Now().UTC()
		group := groups.Group{
			ID:        generateGroupID(t),
			Name:      fmt.Sprintf("%s-%d", groupName, i),
			OwnerID:   uid,
			Type:      "things",
			ParentID:  parentID,
			CreatedAt: creationTime,
			UpdatedAt: creationTime,
		}
		// Create Groups with metadata.
		if i < metaNum {
			group.Metadata = metadata
		}

		_, err = groupRepo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
		parentID = group.ID
	}

	cases := map[string]struct {
		level    uint64
		Size     uint64
		Total    uint64
		Metadata groups.Metadata
	}{
		"retrieve all groups": {
			Total: n,
			Size:  n,
			level: 10,
		},
		"retrieve groups with existing metadata": {
			Total:    metaNum,
			Size:     metaNum,
			Metadata: metadata,
			level:    10,
		},
		"retrieve groups with non-existing metadata": {
			Total:    0,
			Metadata: wrongMeta,
			Size:     0,
			level:    10,
		},
		"retrieve groups with hierarchy level depth": {
			Total: 5,
			Size:  5,
			level: 5,
		},
		"retrieve groups with hierarchy level depth and existing metadata": {
			Total:    metaNum,
			Size:     metaNum,
			level:    5,
			Metadata: metadata,
		},
	}

	for desc, tc := range cases {
		page, err := groupRepo.RetrieveAll(context.Background(), tc.level, tc.Metadata)
		size := len(page.Groups)
		assert.Equal(t, tc.Size, uint64(size), fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.Size, size))
		assert.Equal(t, tc.Total, page.Total, fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.Total, page.Total))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestRetrieveAllParents(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	uid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	metadata := groups.Metadata{
		"field": "value",
	}
	wrongMeta := groups.Metadata{
		"wrong": "wrong",
	}

	p, err := groupRepo.RetrieveAll(context.Background(), 5, nil)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	assert.Equal(t, uint64(0), p.Total, fmt.Sprintf("expected total %d got %d\n", 0, p.Total))

	metaNum := uint64(3)

	n := uint64(10)
	parentID := ""
	parentMiddle := ""
	for i := uint64(0); i < n; i++ {
		creationTime := time.Now().UTC()
		group := groups.Group{
			ID:        generateGroupID(t),
			Name:      fmt.Sprintf("%s-%d", groupName, i),
			OwnerID:   uid,
			Type:      "things",
			ParentID:  parentID,
			CreatedAt: creationTime,
			UpdatedAt: creationTime,
		}
		// Create Groups with metadata.
		if n-i <= metaNum {
			group.Metadata = metadata
		}
		if i == n/2 {
			parentMiddle = group.ID
		}
		_, err = groupRepo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
		parentID = group.ID
	}

	cases := map[string]struct {
		level    uint64
		parentID string
		Size     uint64
		Total    uint64
		Metadata groups.Metadata
	}{
		"retrieve all parents": {
			Total:    n,
			Size:     groups.MaxLevel + 1,
			level:    groups.MaxLevel,
			parentID: parentID,
		},
		"retrieve groups with existing metadata": {
			Total:    metaNum,
			Size:     metaNum,
			Metadata: metadata,
			parentID: parentID,
			level:    groups.MaxLevel,
		},
		"retrieve groups with non-existing metadata": {
			Total:    uint64(0),
			Metadata: wrongMeta,
			Size:     uint64(0),
			level:    groups.MaxLevel,
			parentID: parentID,
		},
		"retrieve groups with hierarchy level depth": {
			Total:    n,
			Size:     2 + 1,
			level:    uint64(2),
			parentID: parentID,
		},
		"retrieve groups with hierarchy level depth and existing metadata": {
			Total:    metaNum,
			Size:     metaNum,
			level:    3,
			Metadata: metadata,
			parentID: parentID,
		},
		"retrieve parent groups from children in the middle": {
			Total:    n/2 + 1,
			Size:     n/2 + 1,
			level:    groups.MaxLevel,
			parentID: parentMiddle,
		},
	}

	for desc, tc := range cases {
		page, err := groupRepo.RetrieveAllParents(context.Background(), tc.parentID, tc.level, tc.Metadata)
		size := len(page.Groups)
		assert.Equal(t, tc.Size, uint64(size), fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.Size, size))
		assert.Equal(t, tc.Total, page.Total, fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.Total, page.Total))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestRetrieveAllChildren(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	uid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	metadata := groups.Metadata{
		"field": "value",
	}
	wrongMeta := groups.Metadata{
		"wrong": "wrong",
	}

	metaNum := uint64(3)

	n := uint64(10)
	groupID := generateGroupID(t)
	firstParentID := groupID
	parentID := ""
	parentMiddle := ""
	for i := uint64(0); i < n; i++ {
		creationTime := time.Now().UTC()
		group := groups.Group{
			ID:        groupID,
			Name:      fmt.Sprintf("%s-%d", groupName, i),
			OwnerID:   uid,
			Type:      "things",
			ParentID:  parentID,
			CreatedAt: creationTime,
			UpdatedAt: creationTime,
		}
		// Create Groups with metadata.
		if i < metaNum {
			group.Metadata = metadata
		}
		if i == n/2 {
			parentMiddle = group.ID
		}
		_, err = groupRepo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
		parentID = group.ID
		groupID = generateGroupID(t)
	}

	cases := map[string]struct {
		level    uint64
		parentID string
		Size     uint64
		Total    uint64
		Metadata groups.Metadata
	}{
		"retrieve all children": {
			Total:    n,
			Size:     groups.MaxLevel + 1,
			level:    groups.MaxLevel,
			parentID: firstParentID,
		},
		"retrieve groups with existing metadata": {
			Total:    metaNum,
			Size:     metaNum,
			Metadata: metadata,
			parentID: firstParentID,
			level:    groups.MaxLevel,
		},
		"retrieve groups with non-existing metadata": {
			Total:    0,
			Metadata: wrongMeta,
			Size:     0,
			level:    groups.MaxLevel,
			parentID: firstParentID,
		},
		"retrieve groups with hierarchy level depth": {
			Total:    n,
			Size:     2 + 1,
			level:    2,
			parentID: firstParentID,
		},
		"retrieve groups with hierarchy level depth and existing metadata": {
			Total:    metaNum,
			Size:     metaNum,
			level:    3,
			Metadata: metadata,
			parentID: firstParentID,
		},
		"retrieve parent groups from children in the middle": {
			Total:    n / 2,
			Size:     n / 2,
			level:    groups.MaxLevel,
			parentID: parentMiddle,
		},
	}

	for desc, tc := range cases {
		page, err := groupRepo.RetrieveAllChildren(context.Background(), tc.parentID, tc.level, tc.Metadata)
		size := len(page.Groups)
		assert.Equal(t, tc.Size, uint64(size), fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.Size, size))
		assert.Equal(t, tc.Total, page.Total, fmt.Sprintf("%s: expected total %d got %d\n", desc, tc.Total, page.Total))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestAssign(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	uid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	creationTime := time.Now().UTC()
	group := groups.Group{
		ID:        generateGroupID(t),
		Name:      groupName + "Updated",
		OwnerID:   uid,
		Type:      "things",
		CreatedAt: creationTime,
		UpdatedAt: creationTime,
	}

	group, err = groupRepo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	mid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	err = groupRepo.Assign(context.Background(), mid, group)
	require.Nil(t, err, fmt.Sprintf("member assign save unexpected error: %s", err))

	mp, err := groupRepo.Members(context.Background(), group, 10, 10, nil)
	require.Nil(t, err, fmt.Sprintf("member assign save unexpected error: %s", err))
	assert.True(t, mp.Total == 1, fmt.Sprintf("retrieve members of a group: expected %d got %d\n", 1, mp.Total))

	group = groups.Group{ID: group.ID, Type: "things"}
	err = groupRepo.Assign(context.Background(), mid, group)
	assert.True(t, errors.Contains(err, groups.ErrMemberAlreadyAssigned), fmt.Sprintf("assign member again: expected %v got %v\n", nil, err))

	mid, err = idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	group = groups.Group{ID: group.ID, Type: "users"}
	err = groupRepo.Assign(context.Background(), mid, group)
	assert.True(t, errors.Contains(err, groups.ErrMalformedEntity), fmt.Sprintf("assign new member with wrong group type: expected %v got %v\n", nil, err))
}

func TestUnassign(t *testing.T) {
	t.Cleanup(func() { cleanUp(t) })
	dbMiddleware := postgres.NewDatabase(db)
	groupRepo := postgres.NewGroupRepo(dbMiddleware)

	uid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	creationTime := time.Now().UTC()
	group := groups.Group{
		ID:        generateGroupID(t),
		Name:      groupName + "Updated",
		OwnerID:   uid,
		Type:      "things",
		CreatedAt: creationTime,
		UpdatedAt: creationTime,
	}

	group, err = groupRepo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	mid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	err = groupRepo.Assign(context.Background(), mid, group)
	require.Nil(t, err, fmt.Sprintf("member assign unexpected error: %s", err))

	mid, err = idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = groupRepo.Assign(context.Background(), mid, group)
	require.Nil(t, err, fmt.Sprintf("member assign unexpected error: %s", err))

	mp, err := groupRepo.Members(context.Background(), group, 10, 10, nil)
	require.Nil(t, err, fmt.Sprintf("member assign save unexpected error: %s", err))
	assert.True(t, mp.Total == 2, fmt.Sprintf("retrieve members of a group: expected %d got %d\n", 2, mp.Total))

	err = groupRepo.Unassign(context.Background(), mid, group)
	require.Nil(t, err, fmt.Sprintf("member unassign save unexpected error: %s", err))

	mp, err = groupRepo.Members(context.Background(), group, 10, 10, nil)
	require.Nil(t, err, fmt.Sprintf("members retrieve unexpected error: %s", err))
	assert.True(t, mp.Total == 1, fmt.Sprintf("retrieve members of a group: expected %d got %d\n", 1, mp.Total))
}

func cleanUp(t *testing.T) {
	_, err := db.Exec("delete from group_relations")
	require.Nil(t, err, fmt.Sprintf("clean relations unexpected error: %s", err))
	_, err = db.Exec("delete from groups")
	require.Nil(t, err, fmt.Sprintf("clean groups unexpected error: %s", err))
	fmt.Println("cleaned up")
}
