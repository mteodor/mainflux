package groups

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/internal/groups"
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

func ListMembership(svc groups.Service) endpoint.Endpoint {
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

func CreateGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createGroupReq)
		if err := req.validate(); err != nil {
			return groupRes{}, err
		}

		group := groups.Group{
			Name:        req.Name,
			Description: req.Description,
			ParentID:    req.ParentID,
			Metadata:    req.Metadata,
		}

		gp, err := svc.CreateGroup(ctx, req.token, group)
		if err != nil {
			return groupRes{}, errors.Wrap(groups.ErrCreateGroup, err)
		}
		return groupRes{created: true, ID: gp.ID}, nil
	}
}

func ListGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, errors.Wrap(groups.ErrMalformedEntity, err)
		}
		gp, err := svc.ListGroups(ctx, req.token, req.offset, req.limit, req.metadata)
		if err != nil {
			return groupPageRes{}, errors.Wrap(groups.ErrFetchGroups, err)
		}
		return buildGroupsResponse(gp), nil
	}
}

func ListGroupChildrenEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, errors.Wrap(groups.ErrMalformedEntity, err)
		}
		gp, err := svc.ListChildren(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
		if err != nil {
			return groupPageRes{}, errors.Wrap(groups.ErrFetchGroups, err)
		}
		return buildGroupsResponse(gp), nil
	}
}

func ListGroupParentsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, errors.Wrap(groups.ErrMalformedEntity, err)
		}
		gp, err := svc.ListParents(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
		if err != nil {
			return groupPageRes{}, errors.Wrap(groups.ErrFetchGroups, err)
		}
		return buildGroupsResponse(gp), nil
	}
}

func DeleteGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(groups.ErrMalformedEntity, err)
		}
		if err := svc.RemoveGroup(ctx, req.token, req.groupID); err != nil {
			return nil, errors.Wrap(groups.ErrDeleteGroup, err)
		}
		return groupDeleteRes{}, nil
	}
}

func AssignEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(memberGroupReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(groups.ErrMalformedEntity, err)
		}
		if err := svc.Assign(ctx, req.token, req.memberID, req.groupID); err != nil {
			return nil, errors.Wrap(groups.ErrAssignToGroup, err)
		}
		return assignMemberToGroupRes{}, nil
	}
}

func UnassignEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(memberGroupReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(groups.ErrMalformedEntity, err)
		}
		if err := svc.Unassign(ctx, req.token, req.memberID, req.groupID); err != nil {
			return nil, errors.Wrap(groups.ErrUnasignFromGroup, err)
		}
		return removeMemberFromGroupRes{}, nil
	}
}

func ListMembersEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listMemberGroupReq)
		if err := req.validate(); err != nil {
			return memberPageRes{}, errors.Wrap(groups.ErrMalformedEntity, err)
		}
		mp, err := svc.ListMembers(ctx, req.token, req.groupID, req.offset, req.limit, req.metadata)
		if err != nil {
			return memberPageRes{}, err
		}
		return buildUsersResponse(mp), nil
	}
}

func UpdateGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateGroupReq)
		if err := req.validate(); err != nil {
			return groupRes{}, errors.Wrap(groups.ErrMalformedEntity, err)
		}

		group := groups.Group{
			ID:          req.ID,
			Name:        req.Name,
			Description: req.Description,
			ParentID:    req.ParentID,
			Metadata:    req.Metadata,
		}

		_, err := svc.UpdateGroup(ctx, req.token, group)
		if err != nil {
			return groupRes{}, errors.Wrap(groups.ErrUpdateGroup, err)
		}

		res := groupRes{created: false}
		return res, nil
	}
}

func ViewGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupReq)
		if err := req.validate(); err != nil {
			return viewGroupRes{}, errors.Wrap(groups.ErrMalformedEntity, err)
		}

		g, err := svc.ViewGroup(ctx, req.token, req.groupID)
		if err != nil {
			return viewGroupRes{}, errors.Wrap(groups.ErrFetchGroups, err)
		}
		res := viewGroupRes{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Metadata:    g.Metadata,
			ParentID:    g.ParentID,
			OwnerID:     g.OwnerID,
		}
		return res, nil
	}
}

// endpoint.go
func buildGroupsResponse(gp groups.GroupPage) groupPageRes {
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
			OwnerID:     group.OwnerID,
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

func buildUsersResponse(mp groups.MemberPage) memberPageRes {
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
