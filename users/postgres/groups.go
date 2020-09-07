// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
)

var (
	errSaveGroupDB   = errors.New("save group to DB failed")
	errUpdateGroupDB = errors.New("update group data failed")
	errDeleteGroupDB = errors.New("delete group failed")
	errSelectDb      = errors.New("select thing from db error")

	errFK         = "foreign_key_violation"
	errInvalid    = "invalid_text_representation"
	errTruncation = "string_data_right_truncation"

	groupRegexp = regexp.MustCompile("^[a-zA-Z0-9]+$")
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
	q := `INSERT INTO groups (name, description,  metadata) VALUES (:name, :description,  :metadata) RETURNING id`

	if group.Name == "" || !groupRegexp.MatchString(group.Name) {
		return users.Group{}, users.ErrMalformedEntity
	}

	dbu := toDBGroup(group)
	row, err := gr.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid, errTruncation:
				return users.Group{}, errors.Wrap(users.ErrMalformedEntity, err)
			case errDuplicate:
				return users.Group{}, errors.Wrap(users.ErrGroupConflict, err)
			}
		}

		return users.Group{}, errors.Wrap(users.ErrCreateGroup, err)
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

func (gr groupRepository) Delete(ctx context.Context, groupID string) error {
	tx, err := gr.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(errDeleteGroupDB, err)
	}
	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = errors.Wrap(err, errors.Wrap(errTransRollback, txErr))
			}
			return
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(errDeleteGroupDB, err)
		}
		return
	}()

	params := map[string]interface{}{
		"group": groupID,
	}
	q := `SELECT COUNT(*) FROM group_relations WHERE group_id = :group;`

	tot, err := total(ctx, gr.db, q, params)
	if err != nil {
		return errors.Wrap(errDeleteGroupDB, err)
	}
	if tot > 0 {
		return errors.Wrap(users.ErrDeleteGroupNotEmpty, err)
	}

	qd := `DELETE FROM groups WHERE id = :id`
	dbr := toDBGroup(users.Group{ID: groupID})

	res, err := gr.db.NamedExecContext(ctx, qd, dbr)
	if err != nil {
		return errors.Wrap(errDeleteGroupDB, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(errDeleteGroupDB, err)
	}

	if cnt != 1 {
		return errors.Wrap(users.ErrDeleteGroupMissing, err)
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

func (gr groupRepository) AssignUser(ctx context.Context, userID, groupID string) error {
	tx, err := gr.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(errDeleteGroupDB, err)
	}
	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = errors.Wrap(err, errors.Wrap(errTransRollback, txErr))
			}
			return
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(errDeleteGroupDB, err)
		}
		return
	}()

	q := `SELECT COUNT(*) FROM group_relations WHERE group_id = :group AND user_id = :user ;`
	params := map[string]interface{}{
		"group": groupID,
		"user":  userID,
	}

	tot, err := total(ctx, gr.db, q, params)
	if err != nil {
		return errors.Wrap(errDeleteGroupDB, err)
	}
	if tot > 0 {
		return errors.Wrap(users.ErrUserAlreadyAssigned, err)
	}

	qIns := `INSERT INTO group_relations (group_id, user_id) VALUES (:group, :user)`
	_, err = gr.db.NamedQueryContext(ctx, qIns, params)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		fmt.Println(pqErr.Code.Name())
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid, errTruncation:
				return errors.Wrap(users.ErrMalformedEntity, err)
			case errDuplicate:
				return errors.Wrap(users.ErrGroupConflict, err)
			case errFK:
				return errors.Wrap(users.ErrNotFound, err)
			}
		}

		return errors.Wrap(users.ErrAssignUserToGroup, err)
	}

	return nil
}

func (gr groupRepository) RemoveUser(ctx context.Context, userID, groupID string) error {
	q := `DELETE FROM group_relations WHERE user_id = :user_id AND group_id = :group_id`
	dbr := toDBGroupRelation(userID, groupID)
	if _, err := gr.db.NamedExecContext(ctx, q, dbr); err != nil {
		return errors.Wrap(errUpdateDB, err)
	}
	return nil
}

type dbGroup struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Owner       string     `db:"owner"`
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
	Group string `db:"group_id"`
	User  string `db:"user_id"`
}

func toDBGroupRelation(userID, groupID string) dbGroupRelation {
	return dbGroupRelation{
		Group: groupID,
		User:  userID,
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
