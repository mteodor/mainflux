// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
)

var (
	errSaveUserDB       = errors.New("Save user to DB failed")
	errUpdateDB         = errors.New("Update user email to DB failed")
	errUpdateUserDB     = errors.New("Update user metadata to DB failed")
	errRetrieveDB       = errors.New("Retreiving from DB failed")
	errUpdatePasswordDB = errors.New("Update password to DB failed")
	errMarshal          = errors.New("Failed to marshal metadata")
	errUnmarshal        = errors.New("Failed to unmarshal metadata")
)

var _ users.UserRepository = (*userRepository)(nil)

const errDuplicate = "unique_violation"

type userRepository struct {
	db Database
}

// New instantiates a PostgreSQL implementation of user
// repository.
func NewUserRepo(db Database) users.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (ur userRepository) Save(ctx context.Context, user users.User) error {
	q := `INSERT INTO users (id, email, password, metadata) VALUES (:id, :email, :password, :metadata)`

	dbu, err := toDBUser(user)
	if err != nil {
		return errors.Wrap(errSaveUserDB, err)
	}
	if _, err := ur.db.NamedExecContext(ctx, q, dbu); err != nil {
		return errors.Wrap(errSaveUserDB, err)
	}

	return nil
}

func (ur userRepository) Update(ctx context.Context, user users.User) error {
	q := `UPDATE users SET(email, password, metadata) VALUES (:email, :password, :metadata) WHERE email = :email`

	dbu, err := toDBUser(user)
	if err != nil {
		return errors.Wrap(errUpdateDB, err)
	}
	if _, err := ur.db.NamedExecContext(ctx, q, dbu); err != nil {
		return errors.Wrap(errUpdateDB, err)
	}

	return nil
}

func (ur userRepository) UpdateUser(ctx context.Context, user users.User) error {
	q := `UPDATE users SET metadata = :metadata WHERE email = :email`

	dbu, err := toDBUser(user)
	if err != nil {
		return errors.Wrap(errUpdateUserDB, err)
	}
	if _, err := ur.db.NamedExecContext(ctx, q, dbu); err != nil {
		return errors.Wrap(errUpdateUserDB, err)
	}

	return nil
}

func (ur userRepository) RetrieveByEmail(ctx context.Context, email string, groups bool) (users.User, error) {
	// TO-DO retrieve groups for user if groups TRUE
	q := `SELECT id, password, metadata FROM users WHERE email = $1`

	dbu := dbUser{
		Email: email,
	}
	if err := ur.db.QueryRowxContext(ctx, q, email).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return users.User{}, errors.Wrap(users.ErrNotFound, err)

		}
		return users.User{}, errors.Wrap(errRetrieveDB, err)
	}

	return toUser(dbu)
}

func (ur userRepository) RetrieveByID(ctx context.Context, id string, groups bool) (users.User, error) {
	// TO-DO retrieve groups for user if groups TRUE
	q := `SELECT id, password, metadata FROM users WHERE id = $1`

	dbu := dbUser{
		ID: id,
	}
	if err := ur.db.QueryRowxContext(ctx, q, id).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return users.User{}, errors.Wrap(users.ErrNotFound, err)

		}
		return users.User{}, errors.Wrap(errRetrieveDB, err)
	}

	return toUser(dbu)
}

func (ur userRepository) UpdatePassword(ctx context.Context, email, password string) error {
	q := `UPDATE users SET password = :password WHERE email = :email`

	db := dbUser{
		Email:    email,
		Password: password,
	}

	if _, err := ur.db.NamedExecContext(ctx, q, db); err != nil {
		return errors.Wrap(errUpdatePasswordDB, err)
	}

	return nil
}

// dbMetadata type for handling metadata properly in database/sql
type dbMetadata map[string]interface{}

// Scan - Implement the database/sql scanner interface
func (m *dbMetadata) Scan(value interface{}) error {
	if value == nil {
		m = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		m = &dbMetadata{}
		return users.ErrScanMetadata
	}

	if err := json.Unmarshal(b, m); err != nil {
		return err
	}

	return nil
}

// Value Implements valuer
func (m dbMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, err
}

type dbUser struct {
	ID       string `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Metadata []byte `db:"metadata"`
}

func toDBUser(u users.User) (dbUser, error) {
	data := []byte("{}")
	if len(u.Metadata) > 0 {
		b, err := json.Marshal(u.Metadata)
		if err != nil {
			return dbUser{}, errors.Wrap(errMarshal, err)
		}
		data = b
	}

	return dbUser{
		ID:       u.ID,
		Email:    u.Email,
		Password: u.Password,
		Metadata: data,
	}, nil
}

func toUser(dbu dbUser) (users.User, error) {
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(dbu.Metadata), &metadata); err != nil {
		return users.User{}, errors.Wrap(errUnmarshal, err)
	}

	return users.User{
		ID:       dbu.ID,
		Email:    dbu.Email,
		Password: dbu.Password,
		Metadata: metadata,
	}, nil
}
