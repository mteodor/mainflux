package groups

import "github.com/mainflux/mainflux/pkg/errors"

var (
	// ErrUnauthorizedAccess unauthorized access.
	ErrUnauthorizedAccess = errors.New("unauthorized access")

	// ErrMalformedEntity malformed entity.
	ErrMalformedEntity = errors.New("malformed entity")

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

	// ErrNotFound indicates failure to find group.
	ErrNotFound = errors.New("failed to find group")

	// ErrAssignToGroup indicates failure to assign member to a group.
	ErrAssignToGroup = errors.New("failed to assign member to a group")

	// ErrUnasignFromGroup indicates failure to unassign member from a group.
	ErrUnasignFromGroup = errors.New("failed to unassign member from a group")

	// ErrUnsupportedContentType indicates unacceptable or lack of Content-Type
	ErrUnsupportedContentType = errors.New("unsupported content type")

	// ErrFailedDecode indicates failed to decode request body
	ErrFailedDecode = errors.New("failed to decode request body")
)
