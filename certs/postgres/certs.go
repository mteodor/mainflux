// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/bootstrap"
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
	chanFieldsNum     = 3
	connFieldsNum     = 2
	cleanupQuery      = `DELETE FROM channels ch WHERE NOT EXISTS (
						 SELECT channel_id FROM connections c WHERE ch.mainflux_channel = c.channel_id);`
)

var (
	errSaveDB           = errors.New("failed to save bootstrap configuration to database")
	errMarshalChannel   = errors.New("failed to marshal channel into json")
	errUnmarshalChannel = errors.New("failed to unmarshal json to channel")
	errSaveChannels     = errors.New("failed to insert channels to database")
	errSaveConnections  = errors.New("failed to insert connections to database")
	errRemoveUnknown    = errors.New("failed to remove from uknown configurations in database")
	errSaveUnknown      = errors.New("failed to insert into uknown configurations in database")
	errRetrieve         = errors.New("failed to retreive bootstrap configuration from database")
	errUpdate           = errors.New("failed to update bootstrap configuration in database")
	errRemove           = errors.New("failed to remove bootstrap configuration from database")
	errUpdateChannels   = errors.New("failed to update channels in bootstrap configuration database")
	errRemoveChannels   = errors.New("failed to remove channels from bootstrap configuration in database")
	errDisconnectThing  = errors.New("failed to disconnect thing in bootstrap configuration in database")
)

var _ certs.CertsRepository = (*certsRepository)(nil)

type certsRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewCertsRepository instantiates a PostgreSQL implementation of certs
// repository.
func NewCertsRepository(db *sqlx.DB, log logger.Logger) certs.CertsRepository {
	return &certsRepository{db: db, log: log}
}

func (cr certsRepository) Save(cert certs.Cert) (string, error) {
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
			e = bootstrap.ErrConflict
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
	ThingID string    `db:"thing_id"`
	Serial  string    `db:"serial"`
	Expire  time.Time `db:"expire"`
}

func toDBConfig(c certs.Cert) dbCert {
	return dbCert{
		ThingID: c.ThingID,
		Serial:  c.Serial,
		Expire:  c.Expire,
	}
}

func toConfig(dbcrt dbCert) certs.Cert {
	c := certs.Cert{dbcrt.ThingID, dbcrt.Serial, dbcrt.Expire}
	return c
}
