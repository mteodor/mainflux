package groups

import "github.com/mainflux/mainflux/pkg/errors"

var (
	ErrUnauthorizedAccess = errors.New("unauthorized access")
	ErrMalformedEntity    = errors.New("malformed entity")
	ErrGroupConflict      = errors.New("group already exists")
	ErrCreateGroup        = errors.New("failed to create group")
	ErrDeleteGroupMissing = errors.New("failed to delete group")
	ErrNotFound           = errors.New("failed to find group")
	ErrAssignToGroup      = errors.New("failed to assign member to a group")
	ErrConflict           = errors.New("group conflict")
)
