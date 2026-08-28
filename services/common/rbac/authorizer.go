package rbac

import (
	"errors"

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
        "common/rbac/model.conf",
        "common/rbac/policy.csv",
    )
}

var ErrAccessDenied = errors.New("access denied")

type Resource string

const (
	ResourceProduct Resource = "product"
	ResourceCart    Resource = "cart"
	ResourceOrder   Resource = "order"
)

type Action string

const (
	ActionCreate Action = "create"
	ActionView   Action = "view"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"

	ActionAdd    Action = "add"
	ActionRemove Action = "remove"
)

func (a *Authorizer) Authorize(role string, resource Resource, action Action) error {
	ok, err := a.enforcer.Enforce(role, string(resource), string(action))

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

