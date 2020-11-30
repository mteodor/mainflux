package authz

import (
	context "context"
	"errors"

	casbin "github.com/casbin/casbin/v2"
)

var (

	// ErrUnauthorizedAccess represents unauthorized access.
	ErrUnauthorizedAccess = errors.New("unauthorized access")

	// ErrMalformedEntity
	ErrMalformedEntity = errors.New("malformed request")

	// ErrNotFound
	ErrNotFound = errors.New("entity not found")

	// ErrInvalidReq
	ErrInvalidReq = errors.New("invalid request")
)

type Policy struct {
	Subject string
	Object  string
	Action  string
}

type Service interface {
	// AddPolicy creates new policy
	AddPolicy(context.Context, Policy) (bool, error)

	// RemovePolicy removes existing policy
	RemovePolicy(context.Context, Policy) (bool, error)

	// Authorize - checks if request is authorized
	// against saved policies in database.
	Authorize(context.Context, Policy) (bool, error)
}

var _ Service = (*service)(nil)

type service struct {
	enforcer *casbin.SyncedEnforcer
}

// New instantiates the auth service implementation.
func New(e *casbin.SyncedEnforcer) Service {
	return &service{
		enforcer: e,
	}
}

func (svc service) AddPolicy(ctx context.Context, p Policy) (bool, error) {
	return svc.enforcer.AddPolicy(p.Subject, p.Object, p.Action)
}

func (svc service) RemovePolicy(ctx context.Context, p Policy) (bool, error) {
	return svc.enforcer.RemovePolicy(p.Subject, p.Object, p.Action)
}

func (svc service) Authorize(ctx context.Context, p Policy) (bool, error) {
	return svc.enforcer.Enforce(p.Action, p.Object, p.Subject)
}
