package things

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	maxNameSize = 1024
	offsetKey   = "offset"
	limitKey    = "limit"
	nameKey     = "name"
	metadataKey = "metadata"
	contentType = "application/json"

	defOffset = 0
	defLimit  = 10
)

var (
	errUnsupportedContentType = errors.New("unsupported content type")
	errInvalidQueryParams     = errors.New("invalid query params")
	errInvalidLimitParam      = errors.New("invalid limit query param")
	errInvalidOffsetParam     = errors.New("invalid offset query param")

	// ErrUnsupportedContentType indicates unacceptable or lack of Content-Type
	ErrUnsupportedContentType = errors.New("unsupported content type")
	// ErrFailedDecode indicates failed to decode request body
	ErrFailedDecode = errors.New("failed to decode request body")
)

func ListMembership(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listMemberGroupReq)
		if err := req.validate(); err != nil {
			return memberPageRes{}, err
		}
		mp, err := svc.Memberships(ctx, req.token, req.memberID, req.offset, req.limit, req.metadata)
		if err != nil {
			return memberPageRes{}, err
		}
		return buildGroupsResponse(mp), nil
	}
}

func CreateGroupEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createGroupReq)
		if err := req.validate(); err != nil {
			return createGroupRes{}, err
		}

		group := NewGroup()
		group.SetName(req.Name)
		group.SetDescription(req.Description)
		group.SetMetadata(req.Metadata)
		group.SetParentID(req.ParentID)

		gp, err := svc.CreateGroup(ctx, req.token, group)
		if err != nil {
			return createGroupRes{}, err
		}
		return createGroupRes{
			created:     true,
			ID:          gp.ID(),
			ParentID:    gp.ParentID(),
			Description: gp.Description(),
			Metadata:    gp.Metadata(),
			Name:        gp.Name(),
		}, nil
	}
}

func ListGroupsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, err
		}
		gp, err := svc.Groups(ctx, req.token, req.offset, req.limit, req.metadata)
		if err != nil {
			return groupPageRes{}, err
		}
		return buildGroupsResponse(gp), nil
	}
}

func ListGroupChildrenEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, err
		}
		gp, err := svc.Children(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
		if err != nil {
			return groupPageRes{}, err
		}
		return buildGroupsResponse(gp), nil
	}
}

func ListGroupParentsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, err
		}
		gp, err := svc.Parents(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
		if err != nil {
			return groupPageRes{}, err
		}
		return buildGroupsResponse(gp), nil
	}
}

func DeleteGroupEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		if err := svc.RemoveGroup(ctx, req.token, req.groupID); err != nil {
			return nil, err
		}
		return groupDeleteRes{}, nil
	}
}

func AssignMemberToGroup(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(memberGroupReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		if err := svc.Assign(ctx, req.token, req.memberID, req.groupID); err != nil {
			return nil, err
		}
		return assignMemberToGroupRes{}, nil
	}
}

func RemoveMemberFromGroup(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(memberGroupReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		if err := svc.Unassign(ctx, req.token, req.memberID, req.groupID); err != nil {
			return nil, err
		}
		return removeMemberFromGroupRes{}, nil
	}
}

func ListMembersForGroupEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listMemberGroupReq)
		if err := req.validate(); err != nil {
			return memberPageRes{}, err
		}
		mp, err := svc.Members(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
		if err != nil {
			return memberPageRes{}, err
		}
		return buildUsersResponse(mp), nil
	}
}

func UpdateGroupEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateGroupReq)
		if err := req.validate(); err != nil {
			return createGroupRes{}, err
		}

		group := NewGroup()
		group.SetID(req.ID)
		group.SetDescription(req.Description)
		group.SetMetadata(req.Metadata)

		g, err := svc.UpdateGroup(ctx, req.token, group)
		if err != nil {
			return createGroupRes{}, err
		}

		res := createGroupRes{
			Name:        g.Name(),
			Description: g.Description(),
			Metadata:    g.Metadata(),
			created:     false,
		}
		return res, nil
	}
}

func ViewGroupEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupReq)
		if err := req.validate(); err != nil {
			return viewGroupRes{}, err
		}

		g, err := svc.Group(ctx, req.token, req.groupID)
		if err != nil {
			return viewGroupRes{}, err
		}
		res := viewGroupRes{
			Name:        g.Name(),
			Description: g.Description(),
			Metadata:    g.Metadata(),
			ParentID:    g.ParentID(),
		}
		return res, nil
	}
}

func DecodeListGroupsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}
	o, err := readUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := readUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	n, err := readStringQuery(r, nameKey)
	if err != nil {
		return nil, err
	}

	m, err := readMetadataQuery(r, metadataKey)
	if err != nil {
		return nil, err
	}

	req := listGroupsReq{
		token:    r.Header.Get("Authorization"),
		offset:   o,
		limit:    l,
		name:     n,
		metadata: m,
		groupID:  bone.GetValue(r, "groupID"),
	}
	return req, nil
}

func DecodeListMemberGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}
	o, err := readUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := readUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	n, err := readStringQuery(r, nameKey)
	if err != nil {
		return nil, err
	}

	m, err := readMetadataQuery(r, metadataKey)
	if err != nil {
		return nil, err
	}

	req := listMemberGroupReq{
		token:    r.Header.Get("Authorization"),
		groupID:  bone.GetValue(r, "groupID"),
		memberID: bone.GetValue(r, "memberID"),
		offset:   o,
		limit:    l,
		name:     n,
		metadata: m,
	}
	return req, nil
}

func DecodeGroupCreate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}
	var req createGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(ErrFailedDecode, err)
	}
	req.token = r.Header.Get("Authorization")
	return req, nil
}

func DecodeGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := groupReq{
		token:   r.Header.Get("Authorization"),
		groupID: bone.GetValue(r, "groupID"),
		name:    bone.GetValue(r, "name"),
	}

	return req, nil
}

func DecodeMemberGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := memberGroupReq{
		token:    r.Header.Get("Authorization"),
		groupID:  bone.GetValue(r, "groupID"),
		memberID: bone.GetValue(r, "memberID"),
	}
	return req, nil
}

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
			ID:          group.ID(),
			ParentID:    group.ParentID(),
			Name:        group.Name(),
			Description: group.Description(),
			Metadata:    group.Metadata(),
		}
		res.Groups = append(res.Groups, view)
	}
	return res
}

func buildUsersResponse(mp MemberPage) memberPageRes {
	res := memberPageRes{
		pageRes: pageRes{
			Total:  mp.Total,
			Offset: mp.Offset,
			Limit:  mp.Limit,
		},
		Members: []interface{}{},
	}
	for _, m := range mp.Members {
		res.Members = append(res.Members, m)
	}
	return res
}

// responses.go
var (
	_ mainflux.Response = (*memberPageRes)(nil)
	_ mainflux.Response = (*createGroupRes)(nil)
	_ mainflux.Response = (*updateGroupRes)(nil)
	_ mainflux.Response = (*viewGroupRes)(nil)
	_ mainflux.Response = (*groupDeleteRes)(nil)
	_ mainflux.Response = (*assignMemberToGroupRes)(nil)
	_ mainflux.Response = (*removeMemberFromGroupRes)(nil)
)

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}
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

func (res assignMemberToGroupRes) Headers() map[string]string {
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

func readUintQuery(r *http.Request, key string, def uint64) (uint64, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return 0, errInvalidQueryParams
	}

	if len(vals) == 0 {
		return def, nil
	}

	strval := vals[0]
	val, err := strconv.ParseUint(strval, 10, 64)
	if err != nil {
		return 0, errInvalidQueryParams
	}

	return val, nil
}

func readStringQuery(r *http.Request, key string) (string, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return "", errInvalidQueryParams
	}

	if len(vals) == 0 {
		return "", nil
	}

	return vals[0], nil
}

func readMetadataQuery(r *http.Request, key string) (map[string]interface{}, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return nil, errInvalidQueryParams
	}

	if len(vals) == 0 {
		return nil, nil
	}

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(vals[0]), &m)
	if err != nil {
		return nil, errors.Wrap(errInvalidQueryParams, err)
	}

	return m, nil
}
