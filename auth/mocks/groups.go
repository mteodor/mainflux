// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mainflux/mainflux/auth/groups"
)

var _ groups.Repository = (*groupRepositoryMock)(nil)

type ParentID string
type MemberID string
type ChildID string
type GroupID string

type groupRepositoryMock struct {
	mu          sync.Mutex
	groups      map[GroupID]groups.Group
	children    map[ParentID]map[GroupID]groups.Group
	parents     map[ChildID]ParentID
	memberships map[MemberID]map[GroupID]groups.Group
	members     map[GroupID]map[MemberID]MemberID
}

// NewGroupRepository creates in-memory user repository
func NewGroupRepository() groups.Repository {
	return &groupRepositoryMock{
		groups:      make(map[GroupID]groups.Group),
		children:    make(map[ParentID]map[GroupID]groups.Group),
		parents:     make(map[ChildID]ParentID),
		memberships: make(map[MemberID]map[GroupID]groups.Group),
		members:     make(map[GroupID]map[MemberID]MemberID),
	}
}

func (grm *groupRepositoryMock) Save(ctx context.Context, g groups.Group) (groups.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[GroupID(g.ID)]; ok {
		return groups.Group{}, groups.ErrGroupConflict
	}
	path := g.ID

	if g.ParentID != "" {
		parent, ok := grm.groups[GroupID(g.ParentID)]
		if !ok {
			return groups.Group{}, groups.ErrCreateGroup
		}
		if _, ok := grm.children[ParentID(g.ParentID)]; !ok {
			grm.children[ParentID(g.ParentID)] = make(map[GroupID]groups.Group)
		}
		grm.children[ParentID(g.ParentID)][GroupID(g.ID)] = g
		grm.parents[ChildID(g.ID)] = ParentID(g.ParentID)
		path = fmt.Sprintf("%s.%s", parent.Path, path)
	}

	g.Path = path
	g.Level = len(strings.Split(path, "."))

	grm.groups[GroupID(g.ID)] = g
	return g, nil
}

func (grm *groupRepositoryMock) Update(ctx context.Context, g groups.Group) (groups.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	up, ok := grm.groups[GroupID(g.ID)]
	if !ok {
		return groups.Group{}, groups.ErrNotFound
	}
	up.Name = g.Name
	up.Description = g.Description
	up.Metadata = g.Metadata
	up.UpdatedAt = time.Now()

	grm.groups[GroupID(g.ID)] = up
	return g, nil
}

func (grm *groupRepositoryMock) Delete(ctx context.Context, id string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	_, ok := grm.groups[GroupID(id)]
	if !ok {
		return groups.ErrNotFound
	}

	if len(grm.members[GroupID(id)]) > 0 {
		return groups.ErrGroupNotEmpty
	}

	// This is not quite exact, it should go in depth
	for _, ch := range grm.children[ParentID(id)] {
		if len(grm.members[GroupID(ch.ID)]) > 0 {
			return groups.ErrGroupNotEmpty
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

func (grm *groupRepositoryMock) RetrieveByID(ctx context.Context, id string) (groups.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	val, ok := grm.groups[GroupID(id)]
	if !ok {
		return groups.Group{}, groups.ErrNotFound
	}
	return val, nil
}

func (grm *groupRepositoryMock) RetrieveAll(ctx context.Context, level uint64, m groups.Metadata) (groups.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []groups.Group
	for _, g := range grm.groups {
		items = append(items, g)
	}
	return groups.GroupPage{
		Groups: items,
		PageMetadata: groups.PageMetadata{
			Total: uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) Unassign(ctx context.Context, memberID string, group groups.Group) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[GroupID(group.ID)]; !ok {
		return groups.ErrNotFound
	}
	if _, ok := grm.members[GroupID(group.ID)][MemberID(memberID)]; !ok {
		return groups.ErrNotFound
	}
	delete(grm.members[GroupID(group.ID)], MemberID(memberID))
	delete(grm.memberships[MemberID(memberID)], GroupID(group.ID))
	return nil
}

func (grm *groupRepositoryMock) Assign(ctx context.Context, memberID string, group groups.Group) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[GroupID(group.ID)]; !ok {
		return groups.ErrNotFound
	}
	if _, ok := grm.members[GroupID(group.ID)]; !ok {
		grm.members[GroupID(group.ID)] = make(map[MemberID]MemberID)
	}
	if _, ok := grm.memberships[MemberID(memberID)]; !ok {
		grm.memberships[MemberID(memberID)] = make(map[GroupID]groups.Group)
	}

	grm.members[GroupID(group.ID)][MemberID(memberID)] = MemberID(memberID)
	grm.memberships[MemberID(memberID)][GroupID(group.ID)] = grm.groups[GroupID(group.ID)]
	return nil

}

func (grm *groupRepositoryMock) Memberships(ctx context.Context, memberID string, offset, limit uint64, um groups.Metadata) (groups.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []groups.Group

	first := uint64(offset)
	last := first + uint64(limit)

	i := uint64(0)
	for _, g := range grm.memberships[MemberID(memberID)] {
		if i >= first && i < last {
			items = append(items, g)
		}
		i = i + 1
	}

	return groups.GroupPage{
		Groups: items,
		PageMetadata: groups.PageMetadata{
			Limit:  limit,
			Offset: offset,
			Total:  uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) Members(ctx context.Context, group groups.Group, offset, limit uint64, m groups.Metadata) (groups.MemberPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []string
	members, ok := grm.members[GroupID(group.ID)]
	if !ok {
		return groups.MemberPage{}, groups.ErrNotFound
	}

	first := uint64(offset)
	last := first + uint64(limit)

	i := uint64(0)
	for _, g := range members {
		if i >= first && i < last {
			items = append(items, string(g))
		}
		i = i + 1
	}
	return groups.MemberPage{
		Members: items,
		PageMetadata: groups.PageMetadata{
			Total: uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) RetrieveAllParents(ctx context.Context, groupID string, level uint64, m groups.Metadata) (groups.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if groupID == "" {
		return groups.GroupPage{}, nil
	}

	g, ok := grm.groups[GroupID(groupID)]
	if !ok {
		return groups.GroupPage{}, groups.ErrNotFound
	}

	grps := make([]groups.Group, 0)
	grps, err := grm.getParents(grps, g)
	if err != nil {
		return groups.GroupPage{}, err
	}

	return groups.GroupPage{
		Groups: grps,
		PageMetadata: groups.PageMetadata{
			Total: uint64(len(grps)),
			Name:  g.Type,
		},
	}, nil
}

func (grm *groupRepositoryMock) getParents(grps []groups.Group, g groups.Group) ([]groups.Group, error) {
	grps = append(grps, g)
	parentID, ok := grm.parents[ChildID(g.ID)]
	if !ok && parentID == "" {
		return grps, nil
	}
	parent, ok := grm.groups[GroupID(parentID)]
	if !ok {
		panic(fmt.Sprintf("parent with id: %s not found", parentID))
	}
	return grm.getParents(grps, parent)
}

func (grm *groupRepositoryMock) RetrieveAllChildren(ctx context.Context, groupID string, level uint64, um groups.Metadata) (groups.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	g, ok := grm.groups[GroupID(groupID)]
	if !ok {
		return groups.GroupPage{}, nil
	}

	grps := make([]groups.Group, 0)
	grps = append(grps, g)
	for ch := range grm.parents {
		g, ok := grm.groups[GroupID(ch)]
		if !ok {
			panic(fmt.Sprintf("child with id %s not found", ch))
		}
		grps = append(grps, g)
	}

	return groups.GroupPage{
		Groups: grps,
		PageMetadata: groups.PageMetadata{
			Total: uint64(len(grps)),
			Name:  g.Type,
		},
	}, nil
}
