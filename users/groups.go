// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"
)

// User represents a Mainflux user account. Each user is identified given its
// email and password.
type Group struct {
	ID          string
	Name        string
	Description string
	Attributes  map[string]interface{}
	//Policies   map[string]Policy
	Metadata map[string]interface{}
}

// GroupRepository specifies an group persistence API.
type GroupRepository interface {
	// SaveGroup persists the group.
	SaveGroup(ctx context.Context, g Group) error

	// Update updates the user metadata.
	UpdateGroup(ctx context.Context, g Group) error

	// RetrieveGroupByID retrieves user by its unique identifier.
	RetrieveByID(ctx context.Context, id string) (Group, error)

	// RetrieveByName
	RetrieveByName(ctx context.Context, name string) (Group, error)

	// AssignUserGroup adds user to group
	AssignUserGroup(ctx context.Context, u User, g Group) error
}
