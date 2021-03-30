// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"time"
)

type PolicyDef struct {
	ID          string    `json:"id" db:"id"`
	SubjectType string    `json:"subject_type" db:"subject_type"`
	SubjectID   string    `json:"subject_id" db:"subject_id"`
	Actions     string    `json:"actions" db:"actions"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type SubjectPolicy struct {
	ObjectType string    `json:"object_type" db:"object_type"`
	ObjectID   string    `json:"object_id" db:"object_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type PolicyReq struct {
	SubjectType string `json:"subject_type" db:"subject_type"`
	SubjectID   string `json:"subject_id" db:"subject_id"`
	ObjectType  string `json:"object_type" db:"object_type"`
	ObjectID    string `json:"object_id" db:"object_id"`
}

// type Policy struct {
// 	ID          string    `json:"id"`
// 	Subject     string    `json:"subject_type"`
// 	SubjectID   string    `json:"subject_id"`
// 	Object      string    `json:"object_type"`
// 	ObjectID    string    `json:"object_id"`
// 	Actions     []string  `json:"actions"`
// 	Description string    `json:"description"`
// 	CreatedAt   time.Time `json:"created_at"`
// 	UpdatedAt   time.Time `json:"updated_at"`
// }

type PolicyService interface {
	// CreatePolicy creates policy definition
	CreatePolicy(ctx context.Context, token string, p PolicyDef) error

	//AssignPolicy
	AssignPolicy(ctx context.Context, token, subjectID, subjectType, objectID, objectType string)

	// RetrievePolicy retrieves policy for given subject and object
	RetrievePolicy(ctx context.Context, token, subjectID, subjectType, objectID, objectType string)
}

type PolicyRepository interface {
	// Save policy definition
	SavePolicy(ctx context.Context, p PolicyDef) (PolicyDef, error)

	// AssignPolicy

	// RetrievePolicy retrieve policy
	RetrievePolicy(ctx context.Context, p PolicyReq) (map[string]map[string]PolicyDef, error)
}
