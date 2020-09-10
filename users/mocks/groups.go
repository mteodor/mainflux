// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"sync"

	"github.com/mainflux/mainflux/users"
)

var _ users.GroupRepository = (*groupRepositoryMock)(nil)

type groupRepositoryMock struct {
	mu     sync.Mutex
	groups map[string]users.Group
	// Map of "Maps of users assigned to a group" where group is a key
	users        map[string]map[string]users.User
	groupsByUser map[string]map[string]users.Group
	groupsByName map[string]users.Group
}

// NewGroupRepository creates in-memory user repository
func NewGroupRepository() users.GroupRepository {
	return &groupRepositoryMock{
		groups:       make(map[string]users.Group),
		groupsByName: make(map[string]users.Group),
		users:        make(map[string]map[string]users.User),
		groupsByUser: make(map[string]map[string]users.Group),
	}
}

func (grm *groupRepositoryMock) Save(ctx context.Context, g users.Group) (users.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[g.ID]; ok {
		return users.Group{}, users.ErrGroupConflict
	}
	if _, ok := grm.groupsByName[g.Name]; ok {
		return users.Group{}, users.ErrGroupConflict
	}
	grm.groups[g.ID] = g
	grm.groupsByName[g.Name] = g
	return g, nil
}

func (grm *groupRepositoryMock) Delete(ctx context.Context, id string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[id]; !ok {
		return users.ErrNotFound
	}
	delete(grm.groups, id)
	return nil
}

func (grm *groupRepositoryMock) RemoveUser(ctx context.Context, userID, groupID string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[groupID]; !ok {
		return users.ErrNotFound
	}
	delete(grm.users[groupID], userID)
	return nil
}

func (grm *groupRepositoryMock) Update(ctx context.Context, g users.Group) error {
	return nil
}

func (grm *groupRepositoryMock) Remove(ctx context.Context, g users.Group) error {
	return nil
}

func (grm *groupRepositoryMock) RetrieveByID(ctx context.Context, id string) (users.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	val, ok := grm.groups[id]
	if !ok {
		return users.Group{}, users.ErrNotFound
	}
	return val, nil
}

func (grm *groupRepositoryMock) RetrieveByName(ctx context.Context, name string) (users.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var val users.Group
	err := users.ErrNotFound

	for _, g := range grm.groups {
		if g.Name == name {
			val = g
			err = nil
			break
		}
	}
	return val, err
}

func (grm *groupRepositoryMock) AssignUser(ctx context.Context, userID, groupID string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[groupID]; !ok {
		return users.ErrNotFound
	}
	if _, ok := grm.users[groupID]; !ok {
		grm.users[groupID] = make(map[string]users.User)
	}
	if _, ok := grm.groupsByUser[userID]; !ok {
		grm.groupsByUser[userID] = make(map[string]users.Group)
	}

	grm.users[groupID][userID] = users.User{ID: userID}
	grm.groupsByUser[userID][groupID] = users.Group{ID: groupID}
	return nil

}

func (grm *groupRepositoryMock) RetrieveAllForUser(ctx context.Context, userID string, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []users.Group
	groups, ok := grm.groupsByUser[userID]
	if !ok {
		return users.GroupPage{}, users.ErrNotFound
	}
	for _, g := range groups {
		items = append(items, g)
	}
	return users.GroupPage{
		Groups: items,
		PageMetadata: users.PageMetadata{
			Limit:  limit,
			Offset: offset,
			Total:  uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) RetrieveAll(ctx context.Context, groupID string, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []users.Group
	for _, g := range grm.groups {
		items = append(items, g)
	}
	return users.GroupPage{
		Groups: items,
		PageMetadata: users.PageMetadata{
			Limit:  limit,
			Offset: offset,
			Total:  uint64(len(items)),
		},
	}, nil
}
