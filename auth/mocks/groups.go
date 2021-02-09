// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mainflux/mainflux/auth"
)

var _ auth.GroupRepository = (*groupRepositoryMock)(nil)

type ParentID string
type MemberID string
type ChildID string
type GroupID string

type groupRepositoryMock struct {
	mu          sync.Mutex
	groups      map[GroupID]auth.Group
	children    map[ParentID]map[GroupID]auth.Group
	parents     map[ChildID]ParentID
	memberships map[MemberID]map[GroupID]auth.Group
	members     map[GroupID]map[MemberID]MemberID
}

// NewGroupRepository creates in-memory user repository
func NewGroupRepository() auth.GroupRepository {
	return &groupRepositoryMock{
		groups:      make(map[GroupID]auth.Group),
		children:    make(map[ParentID]map[GroupID]auth.Group),
		parents:     make(map[ChildID]ParentID),
		memberships: make(map[MemberID]map[GroupID]auth.Group),
		members:     make(map[GroupID]map[MemberID]MemberID),
	}
}

func (grm *groupRepositoryMock) Save(ctx context.Context, group auth.Group) (auth.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[GroupID(group.ID)]; ok {
		return auth.Group{}, auth.ErrGroupConflict
	}
	path := group.ID

	if group.ParentID != "" {
		parent, ok := grm.groups[GroupID(group.ParentID)]
		if !ok {
			return auth.Group{}, auth.ErrCreateGroup
		}
		if _, ok := grm.children[ParentID(group.ParentID)]; !ok {
			grm.children[ParentID(group.ParentID)] = make(map[GroupID]auth.Group)
		}
		grm.children[ParentID(group.ParentID)][GroupID(group.ID)] = group
		grm.parents[ChildID(group.ID)] = ParentID(group.ParentID)
		path = fmt.Sprintf("%s.%s", parent.Path, path)
	}

	group.Path = path
	group.Level = len(strings.Split(path, "."))

	grm.groups[GroupID(group.ID)] = group
	return group, nil
}

func (grm *groupRepositoryMock) Update(ctx context.Context, group auth.Group) (auth.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	up, ok := grm.groups[GroupID(group.ID)]
	if !ok {
		return auth.Group{}, auth.ErrNotFound
	}
	up.Name = group.Name
	up.Description = group.Description
	up.Metadata = group.Metadata
	up.UpdatedAt = time.Now()

	grm.groups[GroupID(group.ID)] = up
	return up, nil
}

func (grm *groupRepositoryMock) Delete(ctx context.Context, id string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	_, ok := grm.groups[GroupID(id)]
	if !ok {
		return auth.ErrNotFound
	}

	if len(grm.members[GroupID(id)]) > 0 {
		return auth.ErrGroupNotEmpty
	}

	// This is not quite exact, it should go in depth
	for _, ch := range grm.children[ParentID(id)] {
		if len(grm.members[GroupID(ch.ID)]) > 0 {
			return auth.ErrGroupNotEmpty
		}
	}

	// This is not quite exact, it should go in depth
	delete(grm.groups, GroupID(id))
	for _, ch := range grm.children[ParentID(id)] {
		delete(grm.members, GroupID(ch.ID))
	}

	delete(grm.children, ParentID(id))

	return nil

}

func (grm *groupRepositoryMock) RetrieveByID(ctx context.Context, id string) (auth.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	val, ok := grm.groups[GroupID(id)]
	if !ok {
		return auth.Group{}, auth.ErrNotFound
	}
	return val, nil
}

func (grm *groupRepositoryMock) RetrieveAll(ctx context.Context, pm auth.PageMetadata) (auth.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []auth.Group
	for _, g := range grm.groups {
		items = append(items, g)
	}
	return auth.GroupPage{
		Groups: items,
		PageMetadata: auth.PageMetadata{
			Total: uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) Unassign(ctx context.Context, groupID string, memberIDs ...string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[GroupID(groupID)]; !ok {
		return auth.ErrGroupNotFound
	}
	for _, memberID := range memberIDs {
		if _, ok := grm.members[GroupID(groupID)][MemberID(memberID)]; !ok {
			return auth.ErrGroupNotFound
		}
		delete(grm.members[GroupID(groupID)], MemberID(memberID))
		delete(grm.memberships[MemberID(memberID)], GroupID(groupID))
	}
	return nil
}

func (grm *groupRepositoryMock) Assign(ctx context.Context, groupID, groupType string, memberIDs ...string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[GroupID(groupID)]; !ok {
		return auth.ErrGroupNotFound
	}
	for _, memberID := range memberIDs {
		if _, ok := grm.members[GroupID(groupID)]; !ok {
			grm.members[GroupID(groupID)] = make(map[MemberID]MemberID)
		}
		if _, ok := grm.memberships[MemberID(memberID)]; !ok {
			grm.memberships[MemberID(memberID)] = make(map[GroupID]auth.Group)
		}

		grm.members[GroupID(groupID)][MemberID(memberID)] = MemberID(memberID)
		grm.memberships[MemberID(memberID)][GroupID(groupID)] = grm.groups[GroupID(groupID)]
	}
	return nil

}

func (grm *groupRepositoryMock) Memberships(ctx context.Context, memberID string, pm auth.PageMetadata) (auth.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []auth.Group

	first := uint64(pm.Offset)
	last := first + uint64(pm.Limit)

	i := uint64(0)
	for _, g := range grm.memberships[MemberID(memberID)] {
		if i >= first && i < last {
			items = append(items, g)
		}
		i = i + 1
	}

	return auth.GroupPage{
		Groups: items,
		PageMetadata: auth.PageMetadata{
			Limit:  pm.Limit,
			Offset: pm.Offset,
			Total:  uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) Members(ctx context.Context, groupID, groupType string, pm auth.PageMetadata) (auth.MemberPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []string
	members, ok := grm.members[GroupID(groupID)]
	if !ok {
		return auth.MemberPage{}, auth.ErrGroupNotFound
	}

	first := uint64(pm.Offset)
	last := first + uint64(pm.Limit)

	i := uint64(0)
	for _, g := range members {
		if i >= first && i < last {
			items = append(items, string(g))
		}
		i = i + 1
	}
	return auth.MemberPage{
		Members: items,
		PageMetadata: auth.PageMetadata{
			Total: uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) RetrieveAllParents(ctx context.Context, groupID string, pm auth.PageMetadata) (auth.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if groupID == "" {
		return auth.GroupPage{}, nil
	}

	group, ok := grm.groups[GroupID(groupID)]
	if !ok {
		return auth.GroupPage{}, auth.ErrGroupNotFound
	}

	grps := make([]auth.Group, 0)
	grps, err := grm.getParents(grps, group)
	if err != nil {
		return auth.GroupPage{}, err
	}

	return auth.GroupPage{
		Groups: grps,
		PageMetadata: auth.PageMetadata{
			Total: uint64(len(grps)),
		},
	}, nil
}

func (grm *groupRepositoryMock) getParents(grps []auth.Group, group auth.Group) ([]auth.Group, error) {
	grps = append(grps, group)
	parentID, ok := grm.parents[ChildID(group.ID)]
	if !ok && parentID == "" {
		return grps, nil
	}
	parent, ok := grm.groups[GroupID(parentID)]
	if !ok {
		panic(fmt.Sprintf("parent with id: %s not found", parentID))
	}
	return grm.getParents(grps, parent)
}

func (grm *groupRepositoryMock) RetrieveAllChildren(ctx context.Context, groupID string, pm auth.PageMetadata) (auth.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	group, ok := grm.groups[GroupID(groupID)]
	if !ok {
		return auth.GroupPage{}, nil
	}

	grps := make([]auth.Group, 0)
	grps = append(grps, group)
	for ch := range grm.parents {
		g, ok := grm.groups[GroupID(ch)]
		if !ok {
			panic(fmt.Sprintf("child with id %s not found", ch))
		}
		grps = append(grps, g)
	}

	return auth.GroupPage{
		Groups: grps,
		PageMetadata: auth.PageMetadata{
			Total:  uint64(len(grps)),
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}, nil
}
