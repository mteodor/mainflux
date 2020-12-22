// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

// Config defines the options that are used when connecting to a PostgreSQL instance
type Config struct {
	Host        string
	Port        string
	User        string
	Pass        string
	Name        string
	SSLMode     string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
}

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate
// failure.
func Connect(cfg Config) (*sqlx.DB, error) {
	url := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass, cfg.SSLMode, cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert)

	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := migrateDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateDB(db *sqlx.DB) error {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "authn",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS keys (
						id          UUID NOT NULL,
						type        SMALLINT,
						issuer      VARCHAR(254) NOT NULL,
						issued_at   TIMESTAMP NOT NULL,
						expires_at  TIMESTAMP,
						PRIMARY KEY (id, issuer)
                    )`,
				},
				Down: []string{"DROP TABLE IF EXISTS keys"},
			},
			{
				Id: "authn_2",
				Up: []string{
					`ALTER TABLE IF EXISTS keys RENAME COLUMN issuer TO subject`,
					`ALTER TABLE IF EXISTS keys ADD COLUMN IF NOT EXISTS issuer_id UUID NOT NULL`,
					`ALTER TABLE IF EXISTS keys DROP CONSTRAINT keys_pkey`,
					`ALTER TABLE IF EXISTS keys ADD PRIMARY KEY (id, issuer_id)`,
				},
				Down: []string{
					`ALTER TABLE keys RENAME COLUMN subject TO issuer`,
					`ALTER TABLE IF EXISTS keys DROP CONSTRAINT keys_pkey`,
					`ALTER TABLE IF EXISTS keys ADD PRIMARY KEY (id, issuer)`,
					`ALTER TABLE IF EXISTS keys DROP COLUMN issuer_id`,
				},
			},
			// Should the migrations name be changed also?
			{
				Id: "authn_3",
				Up: []string{
					`CREATE extension LTREE`,
					`CREATE TABLE IF NOT EXISTS groups ( 
						id          VARCHAR(254) UNIQUE NOT NULL,
						parent_id   VARCHAR(254), 
						owner_id    VARCHAR(254),
						name        VARCHAR(254) NOT NULL,
						description VARCHAR(1024),
						metadata    JSONB,
						path        LTREE, 
						created_at  TIMESTAMPTZ,
						updated_at  TIMESTAMPTZ,
						PRIMARY KEY (owner_id, path),
						FOREIGN KEY (parent_id) REFERENCES groups (id) ON DELETE CASCADE
				   )`,
					`CREATE TABLE IF NOT EXISTS group_relations (
						member_id VARCHAR(254), NOT NULL,
						group_id VARCHAR(254), NOT NULL,
						created_at  TIMESTAMPTZ,
						updated_at  TIMESTAMPTZ,
						FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
						PRIMARY KEY (member_id, group_id)
				   )`,
					`CREATE INDEX path_gist_idx ON groups USING GIST (path);`,
				},
			},
		},
	}

	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}
