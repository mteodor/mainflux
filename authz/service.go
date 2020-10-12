package authz

import (
	context "context"
	"errors"

	"github.com/casbin/casbin/v2"
)

var (

	// ErrUnauthorizedAccess represents unauthorized access.
	ErrUnauthorizedAccess = errors.New("unauthorized access")

	// ErrMalformedEntity
	ErrMalformedEntity = errors.New("malformed request")
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
	enforcer *casbin.Enforcer
}

// New instantiates the auth service implementation.
func New(e *casbin.Enforcer) Service {
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

	user, err := svc.users.RetrieveByEmail(ctx, userID)
	if err != nil {
		return false, nil
	}
	allow, err := svc.enforcer.Enforce(userID, obj, action)
	if err != nil {
		return false, err
	}
	if allow {
		return true, nil
	}

	gp, _ := svc.groups.RetrieveAllForUser(ctx, user.ID, 0, 10, nil)
	for _, g := range gp.Groups {
		sub := g.Name
		allow, _ := svc.enforcer.Enforce(sub, obj, action)
		if allow == true {
			return true, nil
		}
	}
	return false, nil

	svc.enforcer.Enforce(p.Action, p.Object, p.Subject)
}
