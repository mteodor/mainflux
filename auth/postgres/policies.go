package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/pkg/errors"
)

type policy struct {
	ID          string    `db:"id"`
	Subject     string    `db:"subject_type"`
	SubjectID   string    `db:"subject_id"`
	Object      string    `db:"object_type"`
	ObjectID    string    `db:"object_id"`
	Actions     string    `db:"actions"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (gr groupRepository) SavePolicy(ctx context.Context, p auth.Policy) (auth.Policy, error) {
	// For root group path is initialized with id
	q := `INSERT INTO policy (id, description, subject_type, subject_id, object_type, object_id, actions, created_at, updated_at) 
		  VALUES (:id, :description, :subject, :subject_id, :object, :object_id, :actions :created_at, :updated_at) 
		  RETURNING id, description, subject, subject_id, object, object_id, actions, created_at, updated_at`

	row, err := gr.db.NamedQueryContext(ctx, q, p)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid, errTruncation:
				return auth.Policy{}, errors.Wrap(auth.ErrMalformedEntity, err)
			case errFK:
				return auth.Policy{}, errors.Wrap(auth.ErrCreateGroup, err)
			case errDuplicate:
				return auth.Policy{}, errors.Wrap(auth.ErrGroupConflict, err)
			}
		}

		return auth.Policy{}, errors.Wrap(auth.ErrCreateGroup, errors.New(pqErr.Message))
	}

	defer row.Close()
	row.Next()
	p = auth.Policy{}
	if err := row.StructScan(&p); err != nil {
		return auth.Policy{}, err
	}
	return p, nil
}

func (gr groupRepository) RetrievePolicy(ctx context.Context, p auth.Policy) (map[string]map[string]auth.Policy, error) {
	q := `SELECT id, description, subject_type, subject_id, object_type, object_id, actions, created_at, updated_at FROM policies
		  WHERE subject_type = :subject_type AND subject_id = :subject_id AND object_type = :object_type AND object_id = :object_id`

	pol := policy{
		Subject:   p.Subject,
		SubjectID: p.SubjectID,
		Object:    p.Object,
		ObjectID:  p.ObjectID,
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, pol)
	if err != nil {
		return map[string]map[string]auth.Policy{}, errors.Wrap(auth.ErrFailedToRetrievePolicy, err)
	}
	defer rows.Close()

	items, err := gr.processPolicyRows(rows)
	if err != nil {
		return map[string]map[string]auth.Policy{}, errors.Wrap(auth.ErrFailedToRetrievePolicy, err)
	}

	return items, nil
}

func (gr groupRepository) processPolicyRows(rows *sqlx.Rows) (map[string]map[string]auth.Policy, error) {
	items := map[string]map[string]auth.Policy{}

	for rows.Next() {
		dbPolicy := policy{}
		if err := rows.StructScan(&dbPolicy); err != nil {
			return items, err
		}
		p := auth.Policy{
			Subject:   dbPolicy.Subject,
			SubjectID: dbPolicy.SubjectID,
			ObjectID:  dbPolicy.ObjectID,
			Object:    dbPolicy.Object,
			Actions:   dbPolicy.Actions,
		}
		_, ok := items[p.Subject]
		if !ok {
			items[p.Subject] = make(map[string]auth.Policy)
		}
		items[p.Subject][p.SubjectID] = p
	}

	return items, nil
}
