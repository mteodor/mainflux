package groups

type createGroupReq struct {
	token       string
	Name        string                 `json:"name,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req createGroupReq) validate() error {
	if req.token == "" {
		return ErrUnauthorizedAccess
	}
	if len(req.Name) > maxNameSize || req.Name == "" {
		return ErrMalformedEntity
	}
	return nil
}

type updateGroupReq struct {
	token       string
	ID          string                 `json:"id,omitempty"`
	Description string                 `json:"description,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateGroupReq) validate() error {
	if req.token == "" {
		return ErrUnauthorizedAccess
	}
	if req.ID == "" {
		return ErrMalformedEntity
	}

	return nil
}

type listGroupsReq struct {
	token    string
	offset   uint64
	limit    uint64
	metadata Metadata
	name     string
	groupID  string
}

func (req listGroupsReq) validate() error {
	if req.token == "" {
		return ErrUnauthorizedAccess
	}
	return nil
}

type listMemberGroupReq struct {
	token    string
	offset   uint64
	limit    uint64
	metadata Metadata
	name     string
	groupID  string
	memberID string
}

func (req listMemberGroupReq) validate() error {
	if req.token == "" {
		return ErrUnauthorizedAccess
	}
	if req.groupID == "" && req.memberID == "" {
		return ErrMalformedEntity
	}
	return nil
}

type memberGroupReq struct {
	token    string
	groupID  string
	memberID string
}

func (req memberGroupReq) validate() error {
	if req.token == "" {
		return ErrUnauthorizedAccess
	}
	if req.groupID == "" && req.memberID == "" {
		return ErrMalformedEntity
	}
	return nil
}

type groupReq struct {
	token   string
	groupID string
	name    string
}

func (req groupReq) validate() error {
	if req.token == "" {
		return ErrUnauthorizedAccess
	}
	if req.groupID == "" && req.name == "" {
		return ErrMalformedEntity
	}
	return nil
}
