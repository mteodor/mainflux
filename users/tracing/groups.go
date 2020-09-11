// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package tracing contains middlewares that will add spans
// to existing traces.
package tracing

import (
	"context"

	"github.com/mainflux/mainflux/users"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	assignUser         = "assign_user"
	saveGroupOp        = "save_group_op"
	deleteGroupOp      = "delete_group_op"
	updateGroupOp      = "update_group"
	retrieveGroupByID  = "retrieve_group_by_id"
	retrieveAll        = "retrieve_all_group"
	retrieveByName     = "retrieve_by_name"
	retrieveAllForUser = "retrieve_all_group_for_user"
	removeUser         = "remove_user_from_group"
)

var _ users.GroupRepository = (*groupRepositoryMiddleware)(nil)

type groupRepositoryMiddleware struct {
	tracer opentracing.Tracer
	repo   users.GroupRepository
}

// GroupRepositoryMiddleware tracks request and their latency, and adds spans
// to context.
func GroupRepositoryMiddleware(repo users.GroupRepository, tracer opentracing.Tracer) users.GroupRepository {
	return groupRepositoryMiddleware{
		tracer: tracer,
		repo:   repo,
	}
}
func (grm groupRepositoryMiddleware) Save(ctx context.Context, group users.Group) (users.Group, error) {
	span := createSpan(ctx, grm.tracer, saveGroupOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Save(ctx, group)
}
func (grm groupRepositoryMiddleware) Update(ctx context.Context, group users.Group) error {
	span := createSpan(ctx, grm.tracer, updateGroupOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Update(ctx, group)
}

func (grm groupRepositoryMiddleware) Delete(ctx context.Context, groupID string) error {
	span := createSpan(ctx, grm.tracer, deleteGroupOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Delete(ctx, groupID)
}

func (grm groupRepositoryMiddleware) RetrieveByID(ctx context.Context, id string) (users.Group, error) {
	span := createSpan(ctx, grm.tracer, retrieveGroupByID)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveByID(ctx, id)
}

func (grm groupRepositoryMiddleware) RetrieveByName(ctx context.Context, name string) (users.Group, error) {
	span := createSpan(ctx, grm.tracer, retrieveByName)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveByName(ctx, name)
}

func (grm groupRepositoryMiddleware) RetrieveAll(ctx context.Context, groupID string, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAll)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAll(ctx, groupID, offset, limit, gm)
}

func (grm groupRepositoryMiddleware) RetrieveAllForUser(ctx context.Context, userID string, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAllForUser)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAllForUser(ctx, userID, offset, limit, gm)
}

func (grm groupRepositoryMiddleware) RemoveUser(ctx context.Context, userID, groupID string) error {
	span := createSpan(ctx, grm.tracer, removeUser)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RemoveUser(ctx, userID, groupID)
}

func (grm groupRepositoryMiddleware) AssignUser(ctx context.Context, userID, groupID string) error {
	span := createSpan(ctx, grm.tracer, assignUser)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.AssignUser(ctx, userID, groupID)
}
