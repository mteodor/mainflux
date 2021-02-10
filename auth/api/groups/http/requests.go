package groups

import (
	"regexp"

	"github.com/mainflux/mainflux/auth"
	groups "github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/pkg/errors"
)

var groupRegexp = regexp.MustCompile("^[A-Za-z0-9]+[A-Za-z0-9_-]*$")

type createGroupReq struct {
	token       string
	Name        string                 `json:"name,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req createGroupReq) validate() error {
	if req.token == "" {
		return groups.ErrUnauthorizedAccess
	}
	if len(req.Name) > maxNameSize || req.Name == "" || !groupRegexp.MatchString(req.Name) {
		return errors.Wrap(groups.ErrMalformedEntity, groups.ErrBadGroupName)
	}
	// If the group is root (i.e. no parent) type must specified
	// otherwise group inherits type from the parent.
	if req.ParentID == "" {
		return errors.Wrap(groups.ErrMalformedEntity, groups.ErrMissingGroupType)
	}

	return nil
}

type updateGroupReq struct {
	token       string
	id          string
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateGroupReq) validate() error {
	if req.token == "" {
		return groups.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return groups.ErrMalformedEntity
	}

	return nil
}

type listGroupsReq struct {
	token    string
	id       string
	level    uint64
	tree     bool // If `true` result is JSON that represents groups hierarchy, otherwise JSON just holds array of groups.
	metadata auth.GroupMetadata
}

func (req listGroupsReq) validate() error {
	if req.token == "" {
		return groups.ErrUnauthorizedAccess
	}

	if req.level > auth.MaxLevel {
		return groups.ErrMaxLevelExceeded
	}

	return nil
}

type listMembersReq struct {
	token     string
	id        string
	groupType string
	offset    uint64
	limit     uint64
	tree      bool
	metadata  auth.GroupMetadata
}

func (req listMembersReq) validate() error {
	if req.token == "" {
		return groups.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return groups.ErrMalformedEntity
	}

	return nil
}

type listMembershipReq struct {
	token    string
	id       string
	offset   uint64
	limit    uint64
	tree     bool
	metadata auth.GroupMetadata
}

func (req listMembershipReq) validate() error {
	if req.token == "" {
		return groups.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return groups.ErrMalformedEntity
	}

	return nil
}

type assignMembersGroupReq struct {
	token     string
	groupID   string
	groupType string
	Members   []string `json:"members"`
}

func (req assignMembersGroupReq) validate() error {
	// if req.token == "" {
	// 	return groups.ErrUnauthorizedAccess
	// }

	if req.groupID == "" && len(req.Members) == 0 && req.groupType == "" {
		return groups.ErrMalformedEntity
	}

	return nil
}

type groupReq struct {
	token string
	id    string
}

func (req groupReq) validate() error {
	if req.token == "" {
		return groups.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return groups.ErrMalformedEntity
	}

	return nil
}
