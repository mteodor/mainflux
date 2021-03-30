// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"sync"

	"github.com/mainflux/mainflux/auth"
)

var _ auth.PolicyRepository = (*policyRepositoryMock)(nil)

type policyRepositoryMock struct {
	mu            sync.Mutex
	policies      map[string]auth.PolicyDef
	policiesIndex map[string]map[string]map[string]string
}

// NewGroupRepository creates in-memory user repository
func NewPolicyRepository() auth.PolicyRepository {
	return &policyRepositoryMock{
		policies:      make(map[string]auth.PolicyDef),
		policiesIndex: make(map[string]map[string]map[string]string),
	}
}

// CreatePolicy creates policy definition
func (pr *policyRepositoryMock) SavePolicy(ctx context.Context, p auth.PolicyDef) (auth.PolicyDef, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if _, ok := pr.policies[p.ID]; ok {
		return auth.PolicyDef{}, auth.ErrConflict
	}
	pr.storeIndex(p)
	pr.policies[p.ID] = p
	return p, nil
}

// RetrievePolicy retrieves policy for given subject and object
func (pr *policyRepositoryMock) RetrievePolicy(ctx context.Context, pReq auth.PolicyReq) (map[string]map[string]auth.PolicyDef, error) {
	panic("not implemented")
}

func (pr *policyRepositoryMock) storeIndex(p auth.PolicyDef) {
	if _, ok := pr.policiesIndex[p.SubjectID]; !ok {
		pr.policiesIndex[p.SubjectID] = make(map[string]map[string]string)
	}
	if _, ok := pr.policiesIndex[p.SubjectID][p.SubjectType]; !ok {
		pr.policiesIndex[p.SubjectType][p.SubjectType] = make(map[string]string)
	}
	pr.policiesIndex[p.SubjectType][p.SubjectType][p.ID] = p.ID
}

func (pr *policyRepositoryMock) searchIndex(p auth.PolicyDef) (map[string]string, error) {
	if _, ok := pr.policiesIndex[p.SubjectID]; !ok {
		return map[string]string{}, auth.ErrNotFound
	}
	ids, ok := pr.policiesIndex[p.SubjectID][p.SubjectType]
	if !ok {
		return map[string]string{}, auth.ErrNotFound
	}
	return ids, nil
}
