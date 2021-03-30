// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package tracing contains middlewares that will add spans to existing traces.
package tracing

import (
	"context"

	"github.com/mainflux/mainflux/auth"
	opentracing "github.com/opentracing/opentracing-go"
)

const ()

var _ auth.PolicyRepository = (*policyRepositoryMiddleware)(nil)

type policyRepositoryMiddleware struct {
	tracer opentracing.Tracer
	repo   auth.PolicyRepository
}

// PolicyRepositoryMiddleware tracks request and their latency, and adds spans to context.
func PolicyRepositoryMiddleware(tracer opentracing.Tracer, pr auth.PolicyRepository) auth.PolicyRepository {
	return policyRepositoryMiddleware{
		tracer: tracer,
		repo:   pr,
	}
}
func (prm policyRepositoryMiddleware) SavePolicy(ctx context.Context, p auth.PolicyDef) (auth.PolicyDef, error) {
	span := createSpan(ctx, prm.tracer, unassign)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return prm.repo.SavePolicy(ctx, p)
}

func (prm policyRepositoryMiddleware) RetrievePolicy(ctx context.Context, p auth.PolicyReq) (map[string]map[string]auth.PolicyDef, error) {
	span := createSpan(ctx, prm.tracer, unassign)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return prm.repo.RetrievePolicy(ctx, p)
}
