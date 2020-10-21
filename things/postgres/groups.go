// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things"
)

var (
	errDeleteGroupDB = errors.New("delete group failed")
	errSelectDb      = errors.New("select group from db error")
)

var _ things.GroupRepository = (*groupRepository)(nil)

type groupRepository struct {
	db Database
}

// NewGroupRepo instantiates a PostgreSQL implementation of group
// repository.
func NewGroupRepo(db Database) things.GroupRepository {
	return &groupRepository{
		db: db,
	}
}

func (gr groupRepository) Save(ctx context.Context, group things.Group) (things.Group, error) {
	var id string
	q := `INSERT INTO thing_groups (name, description, id, owner_id, parent_id, metadata) VALUES (:name, :description, :id, :owner_id, :parent_id, :metadata) RETURNING id`
	if group.ParentID() == "" {
		q = `INSERT INTO thing_groups (name, description, id, owner_id, metadata) VALUES (:name, :description, :id, :owner_id, :metadata) RETURNING id`
	}

	dbu, err := toDBGroup(group)
	if err != nil {
		return nil, err
	}

	row, err := gr.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid, errTruncation:
				return nil, errors.Wrap(things.ErrMalformedEntity, err)
			case errDuplicate:
				return nil, errors.Wrap(things.ErrGroupConflict, err)
			}
		}

		return nil, errors.Wrap(things.ErrCreateGroup, err)
	}

	defer row.Close()
	row.Next()
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	group.SetID(id)
	return group, nil
}

func (gr groupRepository) Update(ctx context.Context, group things.Group) (things.Group, error) {
	q := `UPDATE thing_groups SET(name, description, parent_id, metadata) VALUES (:name, :description, :parent_id, :metadata) WHERE id = :id`

	dbu, err := toDBGroup(group)
	if err != nil {
		return nil, errors.Wrap(errUpdateDB, err)
	}

	if _, err := gr.db.NamedExecContext(ctx, q, dbu); err != nil {
		return nil, errors.Wrap(errUpdateDB, err)
	}

	return group, nil
}

func (gr groupRepository) Delete(ctx context.Context, groupID string) error {
	qd := `DELETE FROM thing_groups WHERE id = :id`
	group := things.NewGroup()
	group.SetID(groupID)
	dbg, err := toDBGroup(group)
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
		return errors.Wrap(things.ErrDeleteGroupMissing, err)
	}
	return nil
}

func (gr groupRepository) RetrieveByID(ctx context.Context, id string) (things.Group, error) {
	q := `SELECT id, name, owner_id, parent_id, description, metadata FROM thing_groups WHERE id = $1`
	dbu := dbGroup{
		ID: id,
	}

	if err := gr.db.QueryRowxContext(ctx, q, id).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(things.ErrNotFound, err)

		}
		return nil, errors.Wrap(errRetrieveDB, err)
	}

	return toGroup(dbu), nil
}

func (gr groupRepository) RetrieveByName(ctx context.Context, name string) (things.Group, error) {
	q := `SELECT id, name, description, metadata FROM thing_groups WHERE name = $1`

	dbu := dbGroup{
		Name: name,
	}

	if err := gr.db.QueryRowxContext(ctx, q, name).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(things.ErrNotFound, err)

		}
		return nil, errors.Wrap(errRetrieveDB, err)
	}

	group := toGroup(dbu)
	return group, nil
}

func (gr groupRepository) RetrieveAllParents(ctx context.Context, groupID string, offset, limit uint64, gm things.Metadata) (things.GroupPage, error) {
	if groupID == "" {
		return things.GroupPage{}, nil
	}

	_, mq, err := getGroupsMetadataQuery("subordinates", gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}
	if mq != "" {
		mq = fmt.Sprintf("WHERE %s", mq)
	}
	sq := `WITH RECURSIVE subordinates AS (
				SELECT id, owner_id, parent_id, name, description, metadata
				FROM thing_groups
				WHERE id = :id
				UNION
					SELECT thing_groups.id, thing_groups.owner_id, thing_groups.parent_id, thing_groups.name, thing_groups.description, thing_groups.metadata
					FROM thing_groups 
					INNER JOIN subordinates s ON s.parent_id = thing_groups.id
			)`
	q := fmt.Sprintf("%s SELECT * FROM subordinates %s ORDER BY id LIMIT :limit OFFSET :offset", sq, mq)
	cq := fmt.Sprintf("%s SELECT COUNT(*) FROM subordinates %s", sq, mq)
	dbPage, err := toDBGroupPage("", groupID, offset, limit, gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}
	defer rows.Close()

	items, err := processRows(rows)
	if err != nil {
		return things.GroupPage{}, err
	}

	total, err := total(ctx, gr.db, cq, dbPage)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	page := things.GroupPage{
		Groups: items,
		PageMetadata: things.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return page, nil
}

func (gr groupRepository) RetrieveAllChildren(ctx context.Context, groupID string, offset, limit uint64, gm things.Metadata) (things.GroupPage, error) {
	if groupID == "" {
		return things.GroupPage{}, nil
	}
	_, mq, err := getGroupsMetadataQuery("subordinates", gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}
	if mq != "" {
		mq = fmt.Sprintf("WHERE %s", mq)
	}
	sq := `WITH RECURSIVE subordinates AS (
				SELECT id, owner_id, parent_id, name, description, metadata
				FROM thing_groups
				WHERE id = :id 
				UNION
					SELECT thing_groups.id, thing_groups.owner_id, thing_groups.parent_id, thing_groups.name, thing_groups.description, thing_groups.metadata
					FROM thing_groups 
					INNER JOIN subordinates s ON s.id = thing_groups.parent_id
			)`
	q := fmt.Sprintf("%s SELECT * FROM subordinates %s ORDER BY id LIMIT :limit OFFSET :offset", sq, mq)
	cq := fmt.Sprintf("%s SELECT COUNT(*) FROM subordinates %s", sq, mq)
	dbPage, err := toDBGroupPage("", groupID, offset, limit, gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}
	defer rows.Close()

	items, err := processRows(rows)
	if err != nil {
		return things.GroupPage{}, err
	}

	total, err := total(ctx, gr.db, cq, dbPage)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	page := things.GroupPage{
		Groups: items,
		PageMetadata: things.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return page, nil
}

func (gr groupRepository) RetrieveAll(ctx context.Context, offset, limit uint64, gm things.Metadata) (things.GroupPage, error) {
	_, mq, err := getGroupsMetadataQuery("thing_groups", gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}
	if mq != "" {
		mq = fmt.Sprintf("WHERE %s", mq)
	}

	cq := fmt.Sprintf("SELECT COUNT(*) FROM thing_groups %s", mq)
	q := fmt.Sprintf("SELECT id, owner_id, parent_id, name, description, metadata FROM thing_groups %s ORDER BY id LIMIT :limit OFFSET :offset", mq)

	dbPage, err := toDBGroupPage("", "", offset, limit, gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}
	defer rows.Close()

	items, err := processRows(rows)
	if err != nil {
		return things.GroupPage{}, err
	}

	total, err := total(ctx, gr.db, cq, dbPage)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	page := things.GroupPage{
		Groups: items,
		PageMetadata: things.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return page, nil
}

func (gr groupRepository) Members(ctx context.Context, groupID string, offset, limit uint64, gm things.Metadata) (things.Page, error) {
	m, mq, err := getGroupsMetadataQuery("things_group", gm)
	if err != nil {
		return things.Page{}, errors.Wrap(errRetrieveDB, err)
	}

	q := fmt.Sprintf(`SELECT th.id, th.name, th.key, th.metadata FROM things th, thing_group_relations g
                      WHERE th.id = g.thing_id AND g.group_id = :group 
                      %s ORDER BY id LIMIT :limit OFFSET :offset;`, mq)

	params := map[string]interface{}{
		"group":    groupID,
		"limit":    limit,
		"offset":   offset,
		"metadata": m,
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return things.Page{}, errors.Wrap(errSelectDb, err)
	}
	defer rows.Close()

	var items []things.Thing
	for rows.Next() {
		dbTh := dbThing{}
		if err := rows.StructScan(&dbTh); err != nil {
			return things.Page{}, errors.Wrap(errSelectDb, err)
		}

		thing, err := toThing(dbTh)
		if err != nil {
			return things.Page{}, err
		}

		items = append(items, thing)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM things th, thing_group_relations g
	WHERE th.id = g.thing_id AND g.group_id = :group  %s;`, mq)

	total, err := total(ctx, gr.db, cq, params)
	if err != nil {
		return things.Page{}, errors.Wrap(errSelectDb, err)
	}

	page := things.Page{
		Things: items,
		PageMetadata: things.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return page, nil
}

func (gr groupRepository) Memberships(ctx context.Context, userID string, offset, limit uint64, gm things.Metadata) (things.GroupPage, error) {
	m, mq, err := getGroupsMetadataQuery("thing_groups", gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}

	if mq != "" {
		mq = fmt.Sprintf("AND %s", mq)
	}
	q := fmt.Sprintf(`SELECT g.id, g.owner_id, g.parent_id, g.name, g.description, g.metadata 
					  FROM thing_group_relations gr, thing_groups g
					  WHERE gr.group_id = g.id and gr.thing_id = :userID 
		  			  %s ORDER BY id LIMIT :limit OFFSET :offset;`, mq)

	params := map[string]interface{}{
		"userID":   userID,
		"limit":    limit,
		"offset":   offset,
		"metadata": m,
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}
	defer rows.Close()

	var items []things.Group
	for rows.Next() {
		dbgr := dbGroup{}
		if err := rows.StructScan(&dbgr); err != nil {
			return things.GroupPage{}, errors.Wrap(errSelectDb, err)
		}
		gr := toGroup(dbgr)
		if err != nil {
			return things.GroupPage{}, err
		}
		items = append(items, gr)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) 
					   FROM thing_group_relations gr, thing_groups g
					   WHERE gr.group_id = g.id and gr.thing_id = :userID %s;`, mq)

	total, err := total(ctx, gr.db, cq, params)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	page := things.GroupPage{
		Groups: items,
		PageMetadata: things.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return page, nil
}

func (gr groupRepository) Assign(ctx context.Context, thingID, groupID string) error {
	dbr, err := toDBGroupRelation(thingID, groupID)
	if err != nil {
		return errors.Wrap(things.ErrAssignToGroup, err)
	}

	qIns := `INSERT INTO thing_group_relations (group_id, thing_id) VALUES (:group_id, :thing_id)`
	_, err = gr.db.NamedQueryContext(ctx, qIns, dbr)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid, errTruncation:
				return errors.Wrap(things.ErrMalformedEntity, err)
			case errDuplicate:
				return errors.Wrap(things.ErrGroupConflict, err)
			case errFK:
				return errors.Wrap(things.ErrNotFound, err)
			}
		}
		return errors.Wrap(things.ErrAssignToGroup, err)
	}

	return nil
}

func (gr groupRepository) Unassign(ctx context.Context, userID, groupID string) error {
	q := `DELETE FROM thing_group_relations WHERE thing_id = :thing_id AND group_id = :group_id`
	dbr, err := toDBGroupRelation(userID, groupID)
	if err != nil {
		return errors.Wrap(things.ErrNotFound, err)
	}
	if _, err := gr.db.NamedExecContext(ctx, q, dbr); err != nil {
		return errors.Wrap(things.ErrConflict, err)
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
	Metadata dbMetadata    `db:"metadata"`
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

func toDBGroup(g things.Group) (dbGroup, error) {
	parentID := ""
	if g.ParentID() != "" {
		parentID = g.ParentID()
	}
	parent, err := toUUID(parentID)
	if err != nil {
		return dbGroup{}, err
	}
	owner, err := toUUID(g.OwnerID())
	if err != nil {
		return dbGroup{}, err
	}

	meta := dbMetadata(g.Metadata())

	return dbGroup{
		ID:          g.ID(),
		Name:        g.Name(),
		ParentID:    parent,
		OwnerID:     owner,
		Description: g.Description(),
		Metadata:    meta,
	}, nil
}

func toDBGroupPage(ownerID, groupID string, offset, limit uint64, metadata things.Metadata) (dbGroupPage, error) {
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
		ID:       group,
		Metadata: dbMetadata(metadata),
		OwnerID:  owner,
		Offset:   offset,
		Limit:    limit,
	}, nil
}

func toGroup(dbu dbGroup) things.Group {
	g := things.NewGroup()
	g.SetID(dbu.ID)
	g.SetName(dbu.Name)
	g.SetParentID(dbu.ParentID.UUID.String())
	g.SetOwnerID(dbu.OwnerID.UUID.String())
	g.SetDescription(dbu.Description)
	meta := things.Metadata(dbu.Metadata)
	g.SetMetadata(meta)
	return g
}

type dbGroupRelation struct {
	Group uuid.UUID `db:"group_id"`
	Thing uuid.UUID `db:"thing_id"`
}

func toDBGroupRelation(thingID, groupID string) (dbGroupRelation, error) {
	group, err := uuid.FromString(groupID)
	if err != nil {
		return dbGroupRelation{}, err
	}
	thing, err := uuid.FromString(thingID)
	if err != nil {
		return dbGroupRelation{}, err
	}
	return dbGroupRelation{
		Group: group,
		Thing: thing,
	}, nil
}

func getGroupsMetadataQuery(db string, m things.Metadata) ([]byte, string, error) {
	mq := ""
	mb := []byte("{}")
	if len(m) > 0 {
		mq = db + `.metadata @> :metadata`

		b, err := json.Marshal(m)
		if err != nil {
			return nil, "", err
		}
		mb = b
	}
	return mb, mq, nil
}

func processRows(rows *sqlx.Rows) ([]things.Group, error) {
	var items []things.Group
	for rows.Next() {
		dbgr := dbGroup{}
		if err := rows.StructScan(&dbgr); err != nil {
			return items, errors.Wrap(errSelectDb, err)
		}
		gr := toGroup(dbgr)

		items = append(items, gr)
	}
	return items, nil
}
