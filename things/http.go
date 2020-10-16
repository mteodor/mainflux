package things

import (
	"context"

	"github.com/mainflux/mainflux/users"
)

const maxNameSize = 1024

// mux.Post("/groups", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "add_group")(things.CreateGroupEndpoint(svc)),
// 	things.DecodeGroupCreate,
// 	encodeResponse,
// 	opts...,
// ))

func CreateGroupEndpoint(svc Service) {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, err
		}
		gp, err := svc.Groups(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
		if err != nil {
			return groupPageRes{}, err
		}
		return buildGroupsResponse(gp), nil
	}
}

// mux.Get("/groups", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "groups")(things.ListGroupsEndpoint(svc)),
// 	things.DecodeListGroups,
// 	encodeResponse,
// 	opts...,
// ))

// mux.Delete("/groups/:groupID", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "delete_group")(things.DeleteGroupEndpoint(svc)),
// 	things.DecodeDeleteGroupRequest,
// 	encodeResponse,
// 	opts...,
// ))

// mux.Put("/groups/:groupID/things/:memberID", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "assign_user_to_group")(things.AssignMemberToGroup(svc)),
// 	things.DecodeMemberGroupRequest,
// 	encodeResponse,
// 	opts...,
// ))

// mux.Delete("/groups/:groupID/things/:memberID", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "remove_thing_from_group")(things.RemoveUserFromGroup(svc)),
// 	things.DecodeMemberGroupRequest,
// 	encodeResponse,
// 	opts...,
// ))

// mux.Get("/groups/:groupID/things", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "members")(things.ListMembersForGroupEndpoint(svc)),
// 	things.DecodeMemberGroupRequest,
// 	encodeResponse,
// 	opts...,
// ))

// mux.Patch("/groups/:groupID", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "update_group")(things.UpdateGroupEndpoint(svc)),
// 	things.DecodeGroupCreate,
// 	encodeResponse,
// 	opts...,
// ))

// mux.Get("/groups/:groupID/groups", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "list_children_groups")(things.ListGroupsEndpoint(svc)),
// 	things.DecodeGroupRequest,
// 	encodeResponse,
// 	opts...,
// ))

// mux.Get("/groups/:groupID", kithttp.NewServer(
// 	kitot.TraceServer(tracer, "group")(things.ViewGroupEndpoint(svc)),
// 	things.DecodeGroupRequest,
// 	encodeResponse,
// 	opts...,
// ))

// endpoint.go
func buildGroupsResponse(gp GroupPage) groupPageRes {
	res := groupPageRes{
		pageRes: pageRes{
			Total:  gp.Total,
			Offset: gp.Offset,
			Limit:  gp.Limit,
		},
		Groups: []viewGroupRes{},
	}
	for _, group := range gp.Groups {
		view := viewGroupRes{
			ID:          group.ID,
			ParentID:    group.ParentID,
			Name:        group.Name,
			Description: group.Description,
			Metadata:    group.Metadata,
		}
		res.Groups = append(res.Groups, view)
	}
	return res
}

func buildUsersResponse(mp MemberPage) userPageRes {
	res := memberPageRes{
		pageRes: pageRes{
			Total:  up.Total,
			Offset: up.Offset,
			Limit:  up.Limit,
		},
		Members: []interface{}{},
	}
	for _, m := range up.Members {
		res.Members = append(res.Members, m)
	}
	return res
}


// responses.go
_ mainflux.Response = (*memberPageRes)(nil)
_ mainflux.Response = (*createGroupRes)(nil)
_ mainflux.Response = (*updateGroupRes)(nil)
_ mainflux.Response = (*viewGroupRes)(nil)
_ mainflux.Response = (*groupDeleteRes)(nil)
_ mainflux.Response = (*assignMemberToGroupRes)(nil)
_ mainflux.Response = (*removeMemberFromGroupRes)(nil)


type memberPageRes struct {
	pageRes
	Members []interface{}
}

func (res memberPageRes) Code() int {
	return http.StatusOK
}

func (res memberPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res memberPageRes) Empty() bool {
	return false
}


type createGroupRes struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	ParentID    string                 `json:"parent_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	created     bool
}

func (res createGroupRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res createGroupRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location": fmt.Sprintf("/groups/%s", res.ID),
		}
	}
	return map[string]string{}
}

func (res createGroupRes) Empty() bool {
	return true
}

type updateGroupRes struct{}

func (res updateGroupRes) Code() int {
	return http.StatusOK
}

func (res updateGroupRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updateGroupRes) Empty() bool {
	return true
}

type viewGroupRes struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	ParentID    string                 `json:"parent_id"`
	OwnerID     string                 `json:"owner_id"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (res viewGroupRes) Code() int {
	return http.StatusOK
}

func (res viewGroupRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewGroupRes) Empty() bool {
	return false
}

type groupPageRes struct {
	pageRes
	Groups []viewGroupRes `json:"groups"`
}

func (res groupPageRes) Code() int {
	return http.StatusOK
}

func (res groupPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res groupPageRes) Empty() bool {
	return false
}

type groupDeleteRes struct{}

func (res groupDeleteRes) Code() int {
	return http.StatusNoContent
}

func (res groupDeleteRes) Headers() map[string]string {
	return map[string]string{}
}

func (res groupDeleteRes) Empty() bool {
	return true
}

type assignMemberToGroupRes struct{}

func (res assignMemberToGroupRes) Code() int {
	return http.StatusNoContent
}

func (res assignUserToGroupRes) Headers() map[string]string {
	return map[string]string{}
}

func (res assignMemberToGroupRes) Empty() bool {
	return true
}

type removeMemberFromGroupRes struct{}

func (res removeMemberFromGroupRes) Code() int {
	return http.StatusNoContent
}

func (res removeMemberFromGroupRes) Headers() map[string]string {
	return map[string]string{}
}

func (res removeMemberFromGroupRes) Empty() bool {
	return true
}

// requests.go

type createGroupReq struct {
	token       string
	Name        string                 `json:"name,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req createGroupReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}
	if len(req.Name) > maxNameSize || req.Name == "" {
		return things.ErrMalformedEntity
	}
	return nil
}

type updateGroupReq struct {
	token       string
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateGroupReq) validate() error {
	if req.token == "" {
		return users.ErrUnauthorizedAccess
	}
	if req.Name == "" {
		return users.ErrMalformedEntity
	}
	if len(req.Name) > maxNameSize {
		return users.ErrMalformedEntity
	}
	return nil
}

type listUserGroupReq struct {
	token    string
	offset   uint64
	limit    uint64
	metadata users.Metadata
	name     string
	groupID  string
	userID   string
}

func (req listUserGroupReq) validate() error {
	if req.token == "" {
		return users.ErrUnauthorizedAccess
	}
	return nil
}

type userGroupReq struct {
	token   string
	groupID string
	userID  string
}

func (req userGroupReq) validate() error {
	if req.token == "" {
		return users.ErrUnauthorizedAccess
	}
	if req.groupID == "" {
		return users.ErrMalformedEntity
	}
	if req.userID == "" {
		return users.ErrMalformedEntity
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
		return users.ErrUnauthorizedAccess
	}
	if req.groupID == "" && req.name == "" {
		return users.ErrMalformedEntity
	}
	return nil
}
