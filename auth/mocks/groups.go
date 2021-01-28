// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
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
	members     map[GroupID]map[MemberID]groups.Member
}

type mockMember struct {
	memberID string
}

func (m mockMember) GetID() string {
	return m.memberID
}

// NewGroupRepository creates in-memory user repository
func NewGroupRepository() groups.Repository {
	return &groupRepositoryMock{
		groups:      make(map[GroupID]groups.Group),
		children:    make(map[ParentID]map[GroupID]groups.Group),
		parents:     make(map[ChildID]ParentID),
		memberships: make(map[MemberID]map[GroupID]groups.Group),
		members:     make(map[GroupID]map[MemberID]groups.Member),
	}
}

func (grm *groupRepositoryMock) Save(ctx context.Context, g groups.Group) (groups.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[GroupID(g.ID)]; ok {
		return groups.Group{}, groups.ErrGroupConflict
	}

	if g.ParentID != "" {
		if _, ok := grm.groups[GroupID(g.ParentID)]; !ok {
			return groups.Group{}, groups.ErrCreateGroup
		}
		if _, ok := grm.children[ParentID(g.ParentID)]; !ok {
			grm.children[ParentID(g.ParentID)] = make(map[GroupID]groups.Group)
		}
		grm.children[ParentID(g.ParentID)][GroupID(g.ID)] = g
		grm.parents[ChildID(g.ID)] = ParentID(g.ParentID)
	}

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
	group, ok := grm.groups[GroupID(id)]
	if !ok {
		return groups.ErrNotFound
	}

	if len(grm.members[GroupID(id)]) > 0 {
		return groups.ErrGroupNotEmpty
	}

	for _, ch := range grm.children[ParentID(id)] {
		if len(grm.members[GroupID(ch.ID)]) > 0 {
			return groups.ErrGroupNotEmpty
		}
	}

	delete(grm.groups, GroupID(id))
	for _, ch := range grm.children[ParentID(id)] {
		delete(grm.members, GroupID(ch.ID))
	}

	delete(grm.children)

	return nil

}

func (grm *groupRepositoryMock) RetrieveByID(ctx context.Context, id string) (groups.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	val, ok := grm.groups[id]
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

func (grm *groupRepositoryMock) Unassign(ctx context.Context, member groups.Member, group groups.Group) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[group.ID]; !ok {
		return groups.ErrNotFound
	}
	delete(grm.members[group.ID], member.GetID())
	delete(grm.groupsByMember, member.GetID())
	return nil
}

func (grm *groupRepositoryMock) Assign(ctx context.Context, member groups.Member, group groups.Group) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[group.ID]; !ok {
		return groups.ErrNotFound
	}
	if _, ok := grm.members[group.ID]; !ok {
		grm.members[group.ID] = make(map[string]groups.Member)
	}

	grm.members[group.ID][member.GetID()] = mockMember{memberID: member.GetID()}
	grm.groupsByMember[member.GetID()][group.ID] = grm.groups[group.ID]
	return nil

}

func (grm *groupRepositoryMock) Memberships(ctx context.Context, memberID string, offset, limit uint64, um groups.Metadata) (groups.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []groups.Group
	for _, g := range grm.memberships[MemberID(memberID)] {
		items = append(items, g)
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
	var items []groups.Member
	members, ok := grm.members[GroupID(group.ID)]
	if !ok {
		return groups.MemberPage{}, groups.ErrNotFound
	}
	for _, g := range members {
		items = append(items, g)
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
		return groups.GroupPage{
			Groups: grps,
		}, err
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
	if !ok {
		return grps, nil
	}
	parent, ok := grm.groups[GroupID(parentID)]
	if !ok {
		return grps, groups.ErrNotFound
	}
	return grm.getParents(grps, parent)
}

func (grm *groupRepositoryMock) RetrieveAllChildren(ctx context.Context, groupID string, level uint64, um groups.Metadata) (groups.GroupPage, error) {
	panic("not implemented")
}
