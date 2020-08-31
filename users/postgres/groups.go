// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
)

var (
	errSaveGroupDB   = errors.New("Save group to DB failed")
	errUpdateGroupDB = errors.New("Update group data failed")
	errSelectDb      = errors.New("select thing from db error")
)

var _ users.GroupRepository = (*groupRepository)(nil)

type groupRepository struct {
	db Database
}

// NewGroupRepo instantiates a PostgreSQL implementation of group
// repository.
func NewGroupRepo(db Database) users.GroupRepository {
	return &groupRepository{
		db: db,
	}
}

func (gr groupRepository) Save(ctx context.Context, group users.Group) (users.Group, error) {
	var id string
	q := `INSERT INTO groups (name, description, owner_id, metadata) VALUES (:name, :description, :owner, :metadata) RETURNING id`

	dbu := toDBGroup(group)
	row, err := gr.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return users.Group{}, errors.Wrap(errSaveGroupDB, err)
	}
	defer row.Close()
	row.Next()
	if err := row.Scan(&id); err != nil {
		return users.Group{}, err
	}
	group.ID = id
	return group, nil
}

func (gr groupRepository) Update(ctx context.Context, group users.Group) error {
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
	return toGroup(dbu), nil
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

func (gr groupRepository) RetrieveAll(ctx context.Context, owner, name string, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
	m, mq, err := getMetadataQuery(gm)
	if err != nil {
		return users.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}

	q := fmt.Sprintf(`SELECT id, owner_id, parent_id, name, description, metadata FROM groups
		  WHERE owner_id = :owner 
		  AND parent_id in (select id from groups where name = :name) 
		  %s ORDER BY id LIMIT :limit OFFSET :offset;`, mq)

	params := map[string]interface{}{
		"owner":    owner,
		"name":     name,
		"limit":    limit,
		"offset":   offset,
		"metadata": m,
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return users.GroupPage{}, errors.Wrap(errSelectDb, err)
	}
	defer rows.Close()

	var items []users.Group
	for rows.Next() {
		dbgr := dbGroup{Owner: owner}
		if err := rows.StructScan(&dbgr); err != nil {
			return users.GroupPage{}, errors.Wrap(errSelectDb, err)
		}

		gr := toGroup(dbgr)
		if err != nil {
			return users.GroupPage{}, err
		}

		items = append(items, gr)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM groups WHERE owner_id = :owner 
	AND parent_id in (select id from groups where name = :name)  %s;`, mq)

	total, err := total(ctx, gr.db, cq, params)
	if err != nil {
		return users.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	page := users.GroupPage{
		Groups: items,
		PageMetadata: users.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return page, nil
}

func (gr groupRepository) RetrieveAllForUser(ctx context.Context, userID string, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
	m, mq, err := getMetadataQuery(gm)
	if err != nil {
		return users.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}

	q := fmt.Sprintf(`SELECT id, owner_id, parent_id, name, description, metadata FROM groups
		  WHERE id IN (SELECT group_id FROM group_relations WHERE user_id = :owner) 
		  %s ORDER BY id LIMIT :limit OFFSET :offset;`, mq)

	params := map[string]interface{}{
		"owner":    userID,
		"limit":    limit,
		"offset":   offset,
		"metadata": m,
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return users.GroupPage{}, errors.Wrap(errSelectDb, err)
	}
	defer rows.Close()

	var items []users.Group
	for rows.Next() {
		dbgr := dbGroup{}
		if err := rows.StructScan(&dbgr); err != nil {
			return users.GroupPage{}, errors.Wrap(errSelectDb, err)
		}

		gr := toGroup(dbgr)
		if err != nil {
			return users.GroupPage{}, err
		}

		items = append(items, gr)
	}
	cq := fmt.Sprintf(`SELECT COUNT(*) FROM groups 
	WHERE id IN (SELECT group_id FROM group_relations WHERE user_id = :owner)  %s;`, mq)

	total, err := total(ctx, gr.db, cq, params)
	if err != nil {
		return users.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	page := users.GroupPage{
		Groups: items,
		PageMetadata: users.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return page, nil
}

func (gr groupRepository) AssignUser(ctx context.Context, u users.User, g users.Group) error {
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
	Owner       string     `db:"owner_id"`
	Parent      string     `db:"parent_id"`
	Description string     `db:"description"`
	Metadata    dbMetadata `db:"metadata"`
}

func toDBGroup(g users.Group) dbGroup {
	return dbGroup{
		ID:          g.ID,
		Name:        g.Name,
		Owner:       g.Owner.ID,
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

func getMetadataQuery(m users.Metadata) ([]byte, string, error) {
	mq := ""
	mb := []byte("{}")
	if len(m) > 0 {
		mq = ` AND users.metadata @> :metadata`

		b, err := json.Marshal(m)
		if err != nil {
			return nil, "", err
		}
		mb = b
	}
	return mb, mq, nil
}

func total(ctx context.Context, db Database, query string, params map[string]interface{}) (uint64, error) {
	rows, err := db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	total := uint64(0)
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, err
		}
	}

	return total, nil
}
