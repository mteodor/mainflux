package groups

import (
	"context"

	"github.com/go-kit/kit/endpoint"
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
		mp, err := svc.ListMemberships(ctx, req.token, req.memberID, req.offset, req.limit, req.metadata)
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

		group := Group{
			Name:        req.Name,
			Description: req.Description,
			ParentID:    req.ParentID,
			Metadata:    req.Metadata,
		}

		gp, err := svc.CreateGroup(ctx, req.token, group)
		if err != nil {
			return createGroupRes{}, err
		}
		return createGroupRes{
			created:     true,
			ID:          gp.ID,
			ParentID:    gp.ParentID,
			Description: gp.Description,
			Metadata:    gp.Metadata,
			Name:        gp.Name,
		}, nil
	}
}

func ListGroupsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, err
		}
		gp, err := svc.ListGroups(ctx, req.token, req.offset, req.limit, req.metadata)
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
		gp, err := svc.ListChildren(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
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
		gp, err := svc.ListParents(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
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
		mp, err := svc.ListMembers(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
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

		group := Group{
			Description: req.Description,
			ParentID:    req.ParentID,
			Metadata:    req.Metadata,
		}

		g, err := svc.UpdateGroup(ctx, req.token, group)
		if err != nil {
			return createGroupRes{}, err
		}

		res := createGroupRes{
			Name:        g.Name,
			Description: g.Description,
			Metadata:    g.Metadata,
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

		g, err := svc.ViewGroup(ctx, req.token, req.groupID)
		if err != nil {
			return viewGroupRes{}, err
		}
		res := viewGroupRes{
			Name:        g.Name,
			Description: g.Description,
			Metadata:    g.Metadata,
			ParentID:    g.ParentID,
		}
		return res, nil
	}
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
			ID:          group.ID,
			ParentID:    group.ParentID,
			Name:        group.Name,
			Description: group.Description,
			Metadata:    group.Metadata,
			Path:        group.Path,
			Level:       group.Level,
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
