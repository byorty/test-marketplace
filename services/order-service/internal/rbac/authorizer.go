package rbac

import (
	"fmt"

	"github.com/casbin/casbin/v2"
)

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
		"internal/rbac/model.conf",
		"internal/rbac/policy.csv",
	)
}

func (a *Authorizer) Authorize(role, resource, action string) error {

	ok, err := a.enforcer.Enforce(role, resource, action)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("permission denied")
	}

	return nil
}