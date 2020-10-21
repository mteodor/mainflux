// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package tracing contains middlewares that will add spans
// to existing traces.
package tracing

import (
	"context"

	"github.com/mainflux/mainflux/things"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	assignMember         = "assign_member"
	saveGroup            = "save_group"
	deleteGroup          = "delete_group"
	updateGroup          = "update_group"
	retrieveGroupByID    = "retrieve_group_by_id"
	retrieveAllAncestors = "retrieve_all_ancestors"
	retrieveAllChildren  = "retrieve_all_children"
	retrieveAll          = "retrieve_all_groups"
	retrieveByName       = "retrieve_by_name"
	memberships          = "memberships"
	members              = "members"
	unassignMember       = "unassign_member"
)

var _ things.GroupRepository = (*groupRepositoryMiddleware)(nil)

type groupRepositoryMiddleware struct {
	tracer opentracing.Tracer
	repo   things.GroupRepository
}

// GroupRepositoryMiddleware tracks request and their latency, and adds spans to context.
func GroupRepositoryMiddleware(tracer opentracing.Tracer, repo things.GroupRepository) things.GroupRepository {
	return groupRepositoryMiddleware{
		tracer: tracer,
		repo:   repo,
	}
}
func (grm groupRepositoryMiddleware) Save(ctx context.Context, group things.Group) (things.Group, error) {
	span := createSpan(ctx, grm.tracer, saveGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Save(ctx, group)
}
func (grm groupRepositoryMiddleware) Update(ctx context.Context, group things.Group) (things.Group, error) {
	span := createSpan(ctx, grm.tracer, updateGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Update(ctx, group)
}

func (grm groupRepositoryMiddleware) Delete(ctx context.Context, groupID string) error {
	span := createSpan(ctx, grm.tracer, deleteGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Delete(ctx, groupID)
}

func (grm groupRepositoryMiddleware) RetrieveByID(ctx context.Context, id string) (things.Group, error) {
	span := createSpan(ctx, grm.tracer, retrieveGroupByID)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveByID(ctx, id)
}

func (grm groupRepositoryMiddleware) RetrieveByName(ctx context.Context, name string) (things.Group, error) {
	span := createSpan(ctx, grm.tracer, retrieveByName)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveByName(ctx, name)
}

func (grm groupRepositoryMiddleware) RetrieveAllParents(ctx context.Context, groupID string, offset, limit uint64, gm things.Metadata) (things.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAllAncestors)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAllParents(ctx, groupID, offset, limit, gm)
}

func (grm groupRepositoryMiddleware) RetrieveAllChildren(ctx context.Context, groupID string, offset, limit uint64, gm things.Metadata) (things.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAllChildren)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAllChildren(ctx, groupID, offset, limit, gm)
}

func (grm groupRepositoryMiddleware) RetrieveAll(ctx context.Context, offset, limit uint64, gm things.Metadata) (things.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAll)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAll(ctx, offset, limit, gm)
}

func (grm groupRepositoryMiddleware) Memberships(ctx context.Context, memberID string, offset, limit uint64, gm things.Metadata) (things.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, memberships)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Memberships(ctx, memberID, offset, limit, gm)
}

func (grm groupRepositoryMiddleware) Members(ctx context.Context, memberID string, offset, limit uint64, gm things.Metadata) (things.Page, error) {
	span := createSpan(ctx, grm.tracer, members)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Members(ctx, memberID, offset, limit, gm)
}

func (grm groupRepositoryMiddleware) Unassign(ctx context.Context, memberID, groupID string) error {
	span := createSpan(ctx, grm.tracer, unassignMember)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Unassign(ctx, memberID, groupID)
}

func (grm groupRepositoryMiddleware) Assign(ctx context.Context, memberID, groupID string) error {
	span := createSpan(ctx, grm.tracer, assignMember)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Assign(ctx, memberID, groupID)
}
