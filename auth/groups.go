package auth

import (
	"context"
	"errors"
	"time"
)

const MaxLevel = 5

type GroupMetadata map[string]interface{}

type Group struct {
	ID          string
	OwnerID     string
	ParentID    string
	Name        string
	Description string
	Metadata    GroupMetadata
	// Indicates a level in hierarchy from first group node.
	// For a root node level is 1.
	Level int
	// Path is a path in a tree, consisted of group names
	// parentName.childrenName1.childrenName2 .
	Path      string
	Type      string
	Children  []*Group
	CreatedAt time.Time
	UpdatedAt time.Time
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

type MemberPage struct {
	PageMetadata
	Members []string
}

type GroupService interface {
	// CreateGroup creates new  group.
	CreateGroup(ctx context.Context, token string, g Group) (Group, error)

	// UpdateGroup updates the group identified by the provided ID.
	UpdateGroup(ctx context.Context, token string, g Group) (Group, error)

	// ViewGroup retrieves data about the group identified by ID.
	ViewGroup(ctx context.Context, token, id string) (Group, error)

	// ListGroups retrieves groups.
	ListGroups(ctx context.Context, token string, level uint64, gm GroupMetadata) (GroupPage, error)

	// ListChildren retrieves groups that are children to group identified by parentID
	ListChildren(ctx context.Context, token, parentID string, level uint64, gm GroupMetadata) (GroupPage, error)

	// ListParents retrieves groups that are parent to group identified by childID.
	ListParents(ctx context.Context, token, childID string, level uint64, gm GroupMetadata) (GroupPage, error)

	// ListMembers retrieves everything that is assigned to a group identified by groupID.
	ListMembers(ctx context.Context, token string, groupID string, offset, limit uint64, gm GroupMetadata) (MemberPage, error)

	// ListMemberships retrieves all groups for member that is identified with memberID belongs to.
	ListMemberships(ctx context.Context, token, memberID string, offset, limit uint64, gm GroupMetadata) (GroupPage, error)

	// RemoveGroup removes the group identified with the provided ID.
	RemoveGroup(ctx context.Context, token, id string) error

	// Assign adds  member with memberID into the group identified by groupID.
	Assign(ctx context.Context, token string, memberID string, groupID string) error

	// Unassign removes member with memberID from group identified by groupID.
	Unassign(ctx context.Context, token string, memberID string, groupID string) error
}

type GroupRepository interface {
	// Save group
	Save(ctx context.Context, g Group) (Group, error)

	// Update a group
	Update(ctx context.Context, g Group) (Group, error)

	// Delete a group
	Delete(ctx context.Context, id string) error

	// RetrieveByID retrieves group by its id
	RetrieveByID(ctx context.Context, id string) (Group, error)

	// RetrieveAll retrieves all groups.
	RetrieveAll(ctx context.Context, level uint64, gm GroupMetadata) (GroupPage, error)

	// RetrieveAllParents retrieves all groups that are ancestors to the group with given groupID.
	RetrieveAllParents(ctx context.Context, groupID string, level uint64, gm GroupMetadata) (GroupPage, error)

	// RetrieveAllChildren retrieves all children from group with given groupID up to the hierarchy level.
	RetrieveAllChildren(ctx context.Context, groupID string, level uint64, gm GroupMetadata) (GroupPage, error)

	//  Retrieves list of groups that member belongs to
	Memberships(ctx context.Context, memberID string, offset, limit uint64, gm GroupMetadata) (GroupPage, error)

	// Members retrieves everything that is assigned to a group identified by groupID.
	Members(ctx context.Context, groupID string, offset, limit uint64, gm GroupMetadata) (MemberPage, error)

	// Assign adds member to group.
	Assign(ctx context.Context, memberID string, groupID string) error

	// Unassign removes a member from a group
	Unassign(ctx context.Context, memberID string, groupID string) error
}

var (
	// ErrMaxLevelExceeded malformed entity.
	ErrMaxLevelExceeded = errors.New("level must be <=5")

	// ErrBadGroupName malformed entity.
	ErrBadGroupName = errors.New("incorrect group name")

	// ErrGroupConflict group conflict.
	ErrGroupConflict = errors.New("group already exists")

	// ErrCreateGroup indicates failure to create group.
	ErrCreateGroup = errors.New("failed to create group")

	// ErrFetchGroups indicates failure to fetch groups.
	ErrFetchGroups = errors.New("failed to fetch groups")

	// ErrUpdateGroup indicates failure to update group.
	ErrUpdateGroup = errors.New("failed to update group")

	// ErrDeleteGroup indicates failure to delete group.
	ErrDeleteGroup = errors.New("failed to delete group")

	// ErrGroupNotFound indicates failure to find group.
	ErrGroupNotFound = errors.New("failed to find group")

	// ErrAssignToGroup indicates failure to assign member to a group.
	ErrAssignToGroup = errors.New("failed to assign member to a group")

	// ErrUnassignFromGroup indicates failure to unassign member from a group.
	ErrUnassignFromGroup = errors.New("failed to unassign member from a group")

	// ErrUnsupportedContentType indicates unacceptable or lack of Content-Type
	ErrUnsupportedContentType = errors.New("unsupported content type")

	// ErrFailedDecode indicates failed to decode request body
	ErrFailedDecode = errors.New("failed to decode request body")

	// ErrMissingParent indicates that parent can't be found
	ErrMissingParent = errors.New("failed to retrieve parent")

	// ErrParentInvariant indicates that parent can't be changed
	ErrParentInvariant = errors.New("parent can't be changed")

	// ErrMissingGroupType indicates missing group type
	ErrMissingGroupType = errors.New("specifying group type is mandatory")

	// ErrInvalidGroupType Invalid group type
	ErrInvalidGroupType = errors.New("invalid group type")

	// ErrGroupNotEmpty indicates group is not empty, can't be deleted.
	ErrGroupNotEmpty = errors.New("group is not empty")

	// ErrMemberAlreadyAssigned indicates that members is already assigned.
	ErrMemberAlreadyAssigned = errors.New("member is already assigned")
)