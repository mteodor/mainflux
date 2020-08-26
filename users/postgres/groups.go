// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
)

var (
	errSaveGroupDB   = errors.New("Save group to DB failed")
	errUpdateGroupDB = errors.New("Update group metadata to DB failed")
)

var _ users.GroupRepository = (*groupRepository)(nil)

type groupRepository struct {
	db Database
}

// New instantiates a PostgreSQL implementation of user
// repository.
func NewGroupRepo(db Database) users.GroupRepository {
	return &groupRepository{
		db: db,
	}
}

func (gr groupRepository) SaveGroup(ctx context.Context, group users.Group) error {
	q := `INSERT INTO groups (id, email, password, metadata) VALUES (:id, :email, :password, :metadata)`

	dbu := toDBGroup(group)
	if _, err := gr.db.NamedExecContext(ctx, q, dbu); err != nil {
		return errors.Wrap(errSaveGroupDB, err)
	}

	return nil
}

func (gr groupRepository) UpdateGroup(ctx context.Context, group users.Group) error {
	q := `UPDATE groups SET(name, description, metadata) VALUES (:name, :description, :metadata) WHERE id = :id`

	dbu := toDBGroup(group)
	if _, err := gr.db.NamedExecContext(ctx, q, dbu); err != nil {
		return errors.Wrap(errUpdateDB, err)
	}

	return nil
}

func (gr groupRepository) RetrieveByID(ctx context.Context, id string) (users.Group, error) {
	q := `SELECT id, name, description, metadata FROM groups WHERE id = $1`

	dbu := dbGroup{
		ID: id,
	}
	if err := gr.db.QueryRowxContext(ctx, q, id).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return users.Group{}, errors.Wrap(users.ErrNotFound, err)

		}
		return users.Group{}, errors.Wrap(errRetrieveDB, err)
	}
	group := toGroup(dbu)
	return group, nil
}

func (gr groupRepository) RetrieveByName(ctx context.Context, name string) (users.Group, error) {
	q := `SELECT id, name, description, metadata FROM groups WHERE name = $1`

	dbu := dbGroup{
		Name: name,
	}
	if err := gr.db.QueryRowxContext(ctx, q, name).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return users.Group{}, errors.Wrap(users.ErrNotFound, err)

		}
		return users.Group{}, errors.Wrap(errRetrieveDB, err)
	}
	group := toGroup(dbu)
	return group, nil
}

func (gr groupRepository) AssignUserGroup(ctx context.Context, u users.User, g users.Group) error {
	q := `INSERT INTO group_relations (group_id, user_id) VALUES (:group_id, :user_id)`
	dbr := toDBGroupRelation(u, g)
	if _, err := gr.db.NamedExecContext(ctx, q, dbr); err != nil {
		return errors.Wrap(errUpdateDB, err)
	}
	return nil
}

type dbGroup struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Metadata    dbMetadata `db:"metadata"`
}

func toDBGroup(g users.Group) dbGroup {
	return dbGroup{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Metadata:    g.Metadata,
	}
}

func toGroup(dbu dbGroup) users.Group {
	return users.Group{
		ID:          dbu.ID,
		Name:        dbu.Name,
		Description: dbu.Description,
		Metadata:    dbu.Metadata,
	}
}

type dbGroupRelation struct {
	GroupID string `db:"group_id"`
	UserID  string `db:"user_id"`
}

func toDBGroupRelation(u users.User, g users.Group) dbGroupRelation {
	return dbGroupRelation{
		GroupID: g.ID,
		UserID:  u.ID,
	}
}
