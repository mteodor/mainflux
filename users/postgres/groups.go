// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
)

var (
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
	var id, q string
	if group.Parent == nil || group.Parent.ID == "" {
		q = `INSERT INTO groups (name, description, id, owner_id, metadata) VALUES (:name, :description, :id, :owner_id, :metadata) RETURNING id`
	} else {
		q = `INSERT INTO groups (name, description, id, owner_id, parent_id, metadata) VALUES (:name, :description, :id, :owner_id, :parent_id, :metadata) RETURNING id`
	}

	if group.ID == "" || group.Name == "" || !groupRegexp.MatchString(group.Name) {
		return users.Group{}, users.ErrMalformedEntity
	}

	dbu, err := toDBGroup(group)
	if err != nil {
		return users.Group{}, err
	}

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

	dbu, err := toDBGroup(group)
	if err != nil {
		return errors.Wrap(errUpdateDB, err)
	}
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
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(errDeleteGroupDB, err)
		}
	}()

	dbr, err := toDBGroupRelation("", groupID)
	if err != nil {
		return errors.Wrap(errDeleteGroupDB, err)
	}
	q := `SELECT COUNT(*) FROM group_relations WHERE group_id = :group_id`

	tot, err := total(ctx, gr.db, q, dbr)
	if err != nil {
		return errors.Wrap(errDeleteGroupDB, err)
	}
	if tot > 0 {
		return errors.Wrap(users.ErrDeleteGroupNotEmpty, err)
	}

	qd := `DELETE FROM groups WHERE id = :id`
	dbg, err := toDBGroup(users.Group{ID: groupID})
	if err != nil {
		return errors.Wrap(errUpdateDB, err)
	}
	res, err := gr.db.NamedExecContext(ctx, qd, dbg)
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

// func (gr groupRepository) RetrieveAll(ctx context.Context, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
// 	_, mq, err := getMetadataQuery(gm)
// 	if err != nil {
// 		return users.GroupPage{}, errors.Wrap(errRetrieveDB, err)
// 	}

// 	q := fmt.Sprintf(`SELECT id, owner_id, parent_id, name, description, metadata FROM groups
// 		  %s ORDER BY id LIMIT :limit OFFSET :offset;`, mq)

// 	dbPage, err := toDBGroupPage("", "", offset, limit, gm)
// 	if err != nil {
// 		return users.GroupPage{}, errors.Wrap(errSelectDb, err)
// 	}
// 	rows, err := gr.db.NamedQueryContext(ctx, q, dbPage)
// 	if err != nil {
// 		return users.GroupPage{}, errors.Wrap(errSelectDb, err)
// 	}
// 	defer rows.Close()

// 	var items []users.Group
// 	for rows.Next() {
// 		dbgr := dbGroup{}
// 		if err := rows.StructScan(&dbgr); err != nil {
// 			return users.GroupPage{}, errors.Wrap(errSelectDb, err)
// 		}

// 		gr := toGroup(dbgr)
// 		if err != nil {
// 			return users.GroupPage{}, err
// 		}

// 		items = append(items, gr)
// 	}

// 	cq := fmt.Sprintf(`SELECT COUNT(*) FROM groups WHERE 1=1 %s;`, mq)

// 	total, err := total(ctx, gr.db, cq, dbPage)
// 	if err != nil {
// 		return users.GroupPage{}, errors.Wrap(errSelectDb, err)
// 	}

// 	page := users.GroupPage{
// 		Groups: items,
// 		PageMetadata: users.PageMetadata{
// 			Total:  total,
// 			Offset: offset,
// 			Limit:  limit,
// 		},
// 	}

// 	return page, nil
// }

func (gr groupRepository) RetrieveAll(ctx context.Context, groupID string, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
	_, mq, err := getMetadataQuery(gm)
	if err != nil {
		return users.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}
	var q string
	parentQuery := ""
	if groupID != "" {
		parentQuery = `AND parent_id in (select id from groups where id = :id)`
	}
	q = fmt.Sprintf(`SELECT id, owner_id, parent_id, name, description, metadata FROM groups WHERE 1=1 %s %s ORDER BY id LIMIT :limit OFFSET :offset;`, parentQuery, mq)

	dbPage, err := toDBGroupPage("", groupID, offset, limit, gm)
	if err != nil {
		return users.GroupPage{}, errors.Wrap(errSelectDb, err)
	}
	rows, err := gr.db.NamedQueryContext(ctx, q, dbPage)
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

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM groups WHERE owner_id = :owner_id
	AND parent_id in (select id from groups where id = :id)  %s;`, mq)

	total, err := total(ctx, gr.db, cq, dbPage)
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
		  WHERE id IN (SELECT group_id FROM group_relations WHERE user_id = :owner_id) 
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
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(errDeleteGroupDB, err)
		}
	}()

	q := `SELECT COUNT(*) FROM group_relations WHERE group_id = :group_id AND user_id = :user_id ;`
	dbr, err := toDBGroupRelation(userID, groupID)
	if err != nil {
		return errors.Wrap(users.ErrAssignUserToGroup, err)
	}
	tot, err := total(ctx, gr.db, q, dbr)
	if err != nil {
		return errors.Wrap(users.ErrAssignUserToGroup, err)
	}
	if tot > 0 {
		return errors.Wrap(users.ErrUserAlreadyAssigned, err)
	}

	qIns := `INSERT INTO group_relations (group_id, user_id) VALUES (:group_id, :user_id)`
	_, err = gr.db.NamedQueryContext(ctx, qIns, dbr)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
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
	dbr, err := toDBGroupRelation(userID, groupID)
	if err != nil {
		return errors.Wrap(users.ErrNotFound, err)
	}
	if _, err := gr.db.NamedExecContext(ctx, q, dbr); err != nil {
		return errors.Wrap(users.ErrConflict, err)
	}
	return nil
}

type dbGroup struct {
	ID          string        `db:"id"`
	Name        string        `db:"name"`
	OwnerID     uuid.NullUUID `db:"owner_id"`
	ParentID    uuid.NullUUID `db:"parent_id"`
	Description string        `db:"description"`
	Metadata    dbMetadata    `db:"metadata"`
}

type dbGroupPage struct {
	ID       uuid.NullUUID `db:"id"`
	OwnerID  uuid.NullUUID `db:"owner_id"`
	ParentID uuid.NullUUID `db:"parent_id"`
	Limit    uint64
	Offset   uint64
	Size     uint64
}

func toUUID(id string) (uuid.NullUUID, error) {
	var parentID uuid.NullUUID
	if err := parentID.Scan(id); err != nil {
		if id != "" {
			return parentID, err
		}
		if err := parentID.Scan(nil); err != nil {
			return parentID, err
		}
	}
	return parentID, nil
}

func toDBGroup(g users.Group) (dbGroup, error) {
	parentID := ""
	if g.Parent != nil && g.Parent.ID != "" {
		parentID = g.Parent.ID
	}
	parent, err := toUUID(parentID)
	if err != nil {
		return dbGroup{}, err
	}
	owner, err := toUUID(g.Owner.ID)
	if err != nil {
		return dbGroup{}, err
	}

	return dbGroup{
		ID:          g.ID,
		Name:        g.Name,
		ParentID:    parent,
		OwnerID:     owner,
		Description: g.Description,
		Metadata:    g.Metadata,
	}, nil
}

func toDBGroupPage(ownerID, groupID string, offset, limit uint64, metadata users.Metadata) (dbGroupPage, error) {
	owner, err := toUUID(ownerID)
	if err != nil {
		return dbGroupPage{}, err
	}
	group, err := toUUID(groupID)
	if err != nil {
		return dbGroupPage{}, err
	}
	if err != nil {
		return dbGroupPage{}, err
	}
	return dbGroupPage{
		ID:      group,
		OwnerID: owner,
		Offset:  offset,
		Limit:   limit,
	}, nil
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
	Group uuid.NullUUID `db:"group_id"`
	User  uuid.NullUUID `db:"user_id"`
}

func toDBGroupRelation(userID, groupID string) (dbGroupRelation, error) {
	group, err := toUUID(groupID)
	if err != nil {
		return dbGroupRelation{}, err
	}
	user, err := toUUID(userID)
	if err != nil {
		return dbGroupRelation{}, err
	}
	return dbGroupRelation{
		Group: group,
		User:  user,
	}, nil
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

func total(ctx context.Context, db Database, query string, params interface{}) (uint64, error) {
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
