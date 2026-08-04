package domain

import "github.com/byorty/test-marketplace/services/rbac"

type Authorizer interface {
    Authorize(role string, resource rbac.Resource, action rbac.Action) error
}