package things

import "context"

type Group interface {
	ID() string
	OwnerID() string
	ParentID() string
	Name() string
	Description() string
	Metadata() Metadata

	SetID(id string)
	SetOwnerID(id string)
	SetParentID(id string)
	SetName(id string)
	SetDescription(id string)
	SetMetadata(m Metadata)
}

type PageMetadata struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Name   string
}

type GroupPage struct {
	PageMetadata
	Groups []Group
}

type Member interface{}

type MemberPage struct {
	PageMetadata
	Members []Member
}

type GroupService interface {
	// CreateGroup creates new  group.
	CreateGroup(ctx context.Context, token string, group Group) (Group, error)

	// UpdateGroup updates the group identified by the provided ID.
	UpdateGroup(ctx context.Context, token string, group Group) error

	// Group retrieves data about the group identified by ID.
	Group(ctx context.Context, token, id string) (Group, error)

	// ListGroups retrieves groups that are children to group identified by parentID
	// if parentID is empty all groups are listed.
	Groups(ctx context.Context, token, parentID string, offset, limit uint64, meta Metadata) (GroupPage, error)

	// Members retrieves everything that is assigned to a group identified by groupID.
	Members(ctx context.Context, token, groupID string, offset, limit uint64, meta Metadata) (MemberPage, error)

	// Memberships retrieves all groups for member that is identified with memberID belongs to.
	Memberships(ctx context.Context, token, memberID string, offset, limit uint64, meta Metadata) (GroupPage, error)

	// RemoveGroup removes the group identified with the provided ID.
	RemoveGroup(ctx context.Context, token, id string) error

	// Assign adds  member with memberID into the group identified by groupID.
	Assign(ctx context.Context, token, memberID, groupID string) error

	// Unassign removes member with memberID from group identified by groupID.
	Unassign(ctx context.Context, token, memberID, groupID string) error
}
