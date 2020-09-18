// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"
)

// Group of users
type Group struct {
	ID          string
	Name        string
	OwnerID     string
	ParentID    string
	Description string
	Metadata    map[string]interface{}
}

// GroupRepository specifies an group persistence API.
type GroupRepository interface {
	// Save persists the group.
	Save(ctx context.Context, g Group) (Group, error)

	// Update updates the group data.
	Update(ctx context.Context, g Group) error

	// Delete deletes group for given id
	Delete(ctx context.Context, id string) error

	// RetrieveByID retrieves group by its unique identifier.
	RetrieveByID(ctx context.Context, id string) (Group, error)

	// RetrieveByName retrieves group by name
	RetrieveByName(ctx context.Context, name string) (Group, error)

	// RetrieveAll retrieves all groups if groupID == "",  if groupID is specified returns children groups
	RetrieveAll(ctx context.Context, groupID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// RetrieveAllForUser retrieves all groups that user belongs to
	RetrieveAllForUser(ctx context.Context, userID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// AssignUser adds user to group.
	AssignUser(ctx context.Context, userID, groupID string) error

	// UnassignUser removes user from group
	UnassignUser(ctx context.Context, userID, groupID string) error
}
