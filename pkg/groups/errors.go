package groups

import "github.com/mainflux/mainflux/pkg/errors"

var (
	ErrUnauthorizedAccess = errors.New("unauthorized access")
	ErrMalformedEntity    = errors.New("malformed entity")
	ErrGroupConflict      = errors.New("group already exists")
	ErrCreateGroup        = errors.New("cannot create group")
	ErrDeleteGroupMissing = errors.New("cannot delete group")
	ErrNotFound           = errors.New("cannot find group")
	ErrAssignToGroup      = errors.New("cannot assign member to a group")
	ErrConflict           = errors.New("group conflict")
)
