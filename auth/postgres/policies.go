// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/pkg/errors"
)

var _ auth.GroupRepository = (*groupRepository)(nil)

type policyRepository struct {
	db Database
}

// NewPolicyRepo instantiates a PostgreSQL implementation of group
// repository.
func NewPolicyRepo(db Database) auth.PolicyRepository {
	return &policyRepository{
		db: db,
	}
}
func (gr policyRepository) SavePolicy(ctx context.Context, p auth.PolicyDef) (auth.PolicyDef, error) {
	// For root group path is initialized with id
	q := `INSERT INTO policy (id, description,  object_type, object_id, actions, created_at, updated_at) 
		  VALUES (:id, :description, :subject, :subject_id, :object, :object_id, :actions :created_at, :updated_at) 
		  RETURNING id, description, subject, subject_id, object, object_id, actions, created_at, updated_at`

	row, err := gr.db.NamedQueryContext(ctx, q, p)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid, errTruncation:
				return auth.PolicyDef{}, errors.Wrap(auth.ErrMalformedEntity, err)
			case errFK:
				return auth.PolicyDef{}, errors.Wrap(auth.ErrCreateGroup, err)
			case errDuplicate:
				return auth.PolicyDef{}, errors.Wrap(auth.ErrGroupConflict, err)
			}
		}

		return auth.PolicyDef{}, errors.Wrap(auth.ErrCreateGroup, errors.New(pqErr.Message))
	}

	defer row.Close()
	row.Next()
	p = auth.PolicyDef{}
	if err := row.StructScan(&p); err != nil {
		return auth.PolicyDef{}, err
	}
	return p, nil
}

func (gr policyRepository) RetrievePolicy(ctx context.Context, pReq auth.PolicyReq) (map[string]map[string]auth.PolicyDef, error) {
	q := `SELECT id, description, subject_type, subject_id, actions, created_at, updated_at FROM policy_def pd, subject_policy sp
		  WHERE pd.id = sp.policy_id AND sp.subject_type = :subject_type OR sp.subject_id = :subject_id AND pd.object_type = :object_type OR pd.object_id = :object_id`

	rows, err := gr.db.NamedQueryContext(ctx, q, pReq)
	if err != nil {
		return map[string]map[string]auth.PolicyDef{}, errors.Wrap(auth.ErrFailedToRetrievePolicy, err)
	}
	defer rows.Close()

	items, err := gr.processPolicyRows(rows)
	if err != nil {
		return map[string]map[string]auth.PolicyDef{}, errors.Wrap(auth.ErrFailedToRetrievePolicy, err)
	}

	return items, nil
}

func (gr policyRepository) processPolicyRows(rows *sqlx.Rows) (map[string]map[string]auth.PolicyDef, error) {
	items := map[string]map[string]auth.PolicyDef{}

	for rows.Next() {
		dbPolicy := auth.PolicyDef{}
		if err := rows.StructScan(&dbPolicy); err != nil {
			return items, err
		}

		_, ok := items[dbPolicy.SubjectType]
		if !ok {
			items[dbPolicy.SubjectType] = make(map[string]auth.PolicyDef)
		}
		items[dbPolicy.SubjectType][dbPolicy.SubjectID] = dbPolicy
	}

	return items, nil
}
