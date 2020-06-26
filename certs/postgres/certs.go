// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	duplicateErr      = "unique_violation"
	uuidErr           = "invalid input syntax for type uuid"
	connConstraintErr = "connections_config_id_fkey"
	fkViolation       = "foreign_key_violation"
	configFieldsNum   = 8
)

var (
	errSaveDB           = errors.New("failed to save certs to database")
	errMarshalChannel   = errors.New("failed to marshal channel into json")
	errUnmarshalChannel = errors.New("failed to unmarshal json to channel")
	errRetrieve         = errors.New("failed to retrieve certs  from database")
	errEmptyThingID     = errors.New("failed to retrieve, thing id empty")
	errUpdate           = errors.New("failed to update certs in database")
	errRemove           = errors.New("failed to remove certs from database")
)

var _ certs.CertsRepository = (*certsRepository)(nil)

type Cert struct {
	ThingID string
	Serial  string
	Expire  time.Duration
}

type certsRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewCertsRepository instantiates a PostgreSQL implementation of certs
// repository.
func NewCertsRepository(db *sqlx.DB, log logger.Logger) certs.CertsRepository {
	return &certsRepository{db: db, log: log}
}

func (cr certsRepository) RetrieveAll(ctx context.Context, thingID string, offset, limit uint64) (certs.CertsPage, error) {
	if thingID == "" {
		return certs.CertsPage{}, errEmptyThingID
	}
	q := `SELECT FROM certs WHERE thing_id = $1 ORDER BY expire LIMIT $2 OFFSET $3;`
	rows, err := cr.db.Query(q, thingID, limit, offset)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve configs due to %s", err))
		return certs.CertsPage{}, err
	}
	defer rows.Close()

	certificates := []certs.Cert{}

	for rows.Next() {
		c := certs.Cert{}
		if err := rows.Scan(&c.ThingID, &c.Serial, &c.Expire); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved config due to %s", err))
			return certs.CertsPage{}, err

		}
		certificates = append(certificates, c)
	}

	q = fmt.Sprintf(`SELECT COUNT(*) FROM certs WHERE thing_id = $1`)
	var total uint64
	if err := cr.db.QueryRow(q, thingID).Scan(&total); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to count certs due to %s", err))
		return certs.CertsPage{}, err
	}

	return certs.CertsPage{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Certs:  certificates,
	}, nil

}

func (cr certsRepository) Save(ctx context.Context, cert certs.Cert) (string, error) {
	q := `INSERT INTO certs (thing_id, serial, expire)
		  VALUES (:thing_id, :serial, :expire)`

	tx, err := cr.db.Beginx()
	if err != nil {
		return "", errors.Wrap(errSaveDB, err)
	}

	dbcrt := toDBConfig(cert)

	if _, err := tx.NamedExec(q, dbcrt); err != nil {
		e := err
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == duplicateErr {
			e = errors.New("error conflict")
		}

		cr.rollback("Failed to insert a Cert", tx, err)

		return "", errors.Wrap(errSaveDB, e)
	}

	if err := tx.Commit(); err != nil {
		cr.rollback("Failed to commit Config save", tx, err)
	}

	return cert.Serial, nil
}

func (cr certsRepository) rollback(content string, tx *sqlx.Tx, err error) {
	cr.log.Error(fmt.Sprintf("%s %s", content, err))

	if err := tx.Rollback(); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to rollback due to %s", err))
	}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

type dbCert struct {
	ThingID string        `db:"thing_id"`
	Serial  string        `db:"serial"`
	Expire  time.Duration `db:"expire"`
}

func toDBConfig(c certs.Cert) dbCert {
	return dbCert{
		ThingID: c.ThingID,
		Serial:  c.Serial,
		Expire:  c.Expire,
	}
}

func toConfig(dbcrt dbCert) Cert {
	c := Cert{
		ThingID: dbcrt.ThingID,
		Serial:  dbcrt.Serial,
		Expire:  dbcrt.Expire,
	}
	return c
}
