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

	// Delete deletes group for given id.
	Delete(ctx context.Context, id string) error

	// RetrieveByID retrieves group by its unique identifier.
	RetrieveByID(ctx context.Context, id string) (Group, error)

	// RetrieveByName retrieves group by name.
	RetrieveByName(ctx context.Context, name string) (Group, error)

	// RetrieveAll retrieves all groups.
	RetrieveAll(ctx context.Context, offset, limit uint64, gm Metadata) (GroupPage, error)

	// RetrieveAllAncestors retrieves all groups that are ancestors to the group with given groupID.
	RetrieveAllAncestors(ctx context.Context, groupID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// RetrieveAllChildren retrieves all children from group with given groupID.
	RetrieveAllChildren(ctx context.Context, groupID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// Memberships retrieves all groups that user belongs to
	Memberships(ctx context.Context, userID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// Assign adds user to group.
	Assign(ctx context.Context, userID, groupID string) error

	// Unassign removes user from group
	Unassign(ctx context.Context, userID, groupID string) error
}
