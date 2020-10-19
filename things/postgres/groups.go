// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/users"
)

var (
	errDeleteGroupDB = errors.New("delete group failed")
	errSelectDb      = errors.New("select group from db error")
)

var _ users.GroupRepository = (*groupRepository)(nil)

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
	if group.ParentID == "" {
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
				return nil, errors.Wrap(users.ErrMalformedEntity, err)
			case errDuplicate:
				return nil, errors.Wrap(users.ErrGroupConflict, err)
			}
		}

		return nil, errors.Wrap(users.ErrCreateGroup, err)
	}

	defer row.Close()
	row.Next()
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	group.SetID(id)
	return group, nil
}

func (gr groupRepository) Update(ctx context.Context, group users.Group) error {
	q := `UPDATE thing_groups SET(name, description, metadata) VALUES (:name, :description, :metadata) WHERE id = :id`

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
	qd := `DELETE FROM thing_groups WHERE id = :id`
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
	q := `SELECT id, name, owner_id, parent_id, description, metadata FROM thing_groups WHERE id = $1`
	dbu := dbGroup{
		ID: id,
	}

	if err := gr.db.QueryRowxContext(ctx, q, id).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(users.ErrNotFound, err)

		}
		return nil, errors.Wrap(errRetrieveDB, err)
	}

	return toGroup(dbu), nil
}

func (gr groupRepository) RetrieveByName(ctx context.Context, name string) (users.Group, error) {
	q := `SELECT id, name, description, metadata FROM thing_groups WHERE name = $1`

	dbu := dbGroup{
		Name: name,
	}

	if err := gr.db.QueryRowxContext(ctx, q, name).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(users.ErrNotFound, err)

		}
		return nil, errors.Wrap(errRetrieveDB, err)
	}

	group := toGroup(dbu)
	return group, nil
}

func (gr groupRepository) RetrieveAllWithAncestors(ctx context.Context, groupID string, offset, limit uint64, gm users.Metadata) (things.GroupPage, error) {
	_, mq, err := getGroupsMetadataQuery(gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}
	if mq != "" {
		mq = fmt.Sprintf("WHERE %s", mq)
	}

	cq := fmt.Sprintf("SELECT COUNT(*) FROM thing_groups %s", mq)
	sq := fmt.Sprintf("SELECT id, owner_id, parent_id, name, description, metadata FROM thing_groups %s", mq)
	q := fmt.Sprintf("%s ORDER BY id LIMIT :limit OFFSET :offset", sq)

	if groupID != "" {
		sq = fmt.Sprintf(
			`WITH RECURSIVE subordinates AS (
				SELECT id, owner_id, parent_id, name, description, metadata
				FROM thing_groups
				WHERE id = :id 
				UNION
					SELECT thing_groups.id, thing_groups.owner_id, thing_groups.parent_id, thing_groups.name, thing_groups.description, thing_groups.metadata
					FROM thing_groups 
					INNER JOIN subordinates s ON s.id = thing_groups.parent_id %s
			)`, mq)
		q = fmt.Sprintf("%s SELECT * FROM subordinates ORDER BY id LIMIT :limit OFFSET :offset", sq)
		cq = fmt.Sprintf("%s SELECT COUNT(*) FROM subordinates", sq)
	}

	dbPage, err := toDBGroupPage("", groupID, offset, limit, gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
	}

	rows, err := gr.db.NamedQueryContext(ctx, q, dbPage)
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

func (gr groupRepository) Memberships(ctx context.Context, userID string, offset, limit uint64, gm users.Metadata) (users.GroupPage, error) {
	m, mq, err := getGroupsMetadataQuery(gm)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errRetrieveDB, err)
	}

	if mq != "" {
		mq = fmt.Sprintf("AND %s", mq)
	}
	q := fmt.Sprintf(`SELECT g.id, g.owner_id, g.parent_id, g.name, g.description, g.metadata 
					  FROM group_relations gr, thing_groups g
					  WHERE gr.group_id = g.id and gr.user_id = :userID 
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

	var items []users.Group
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
					   FROM group_relations gr, thing_groups g
					   WHERE gr.group_id = g.id and gr.user_id = :userID %s;`, mq)

	total, err := total(ctx, gr.db, cq, params)
	if err != nil {
		return things.GroupPage{}, errors.Wrap(errSelectDb, err)
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

func (gr groupRepository) Assign(ctx context.Context, userID, groupID string) error {
	dbr, err := toDBGroupRelation(userID, groupID)
	if err != nil {
		return errors.Wrap(users.ErrAssignUserToGroup, err)
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

func (gr groupRepository) Unassign(ctx context.Context, userID, groupID string) error {
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
	if g.ParentID != "" {
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

func getGroupsMetadataQuery(m things.Metadata) ([]byte, string, error) {
	mq := ""
	mb := []byte("{}")
	if len(m) > 0 {
		mq = `thing_groups.metadata @> :metadata`

		b, err := json.Marshal(m)
		if err != nil {
			return nil, "", err
		}
		mb = b
	}
	return mb, mq, nil
}
