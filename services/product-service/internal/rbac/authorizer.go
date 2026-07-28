package rbac

import (
	"errors"

	"github.com/casbin/casbin/v2"
)

var ErrAccessDenied = errors.New("access denied")

type Authorizer struct {
	enforcer *casbin.Enforcer
}

func New(enforcer *casbin.Enforcer) *Authorizer {
	return &Authorizer{
		enforcer: enforcer,
	}
}

func NewEnforcer() (*casbin.Enforcer, error) {
	return casbin.NewEnforcer(
		"internal/authorization/model.conf",
		"internal/authorization/policy.csv",
	)
}

func (a *Authorizer) Authorize(role, resource, action string) error {

	ok, err := a.enforcer.Enforce(role, resource, action)
	if err != nil {
		return err
	}

	if !ok {
		return ErrAccessDenied
	}

	return nil
}

func IsAccessDenied(err error) bool {
	return errors.Is(err, ErrAccessDenied)
}