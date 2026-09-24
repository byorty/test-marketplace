package mocks

import (
	"sync/atomic"

	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
)

type MockAuthorizer struct {
	AuthorizeFn func(role string, resource rbac.Resource, action rbac.Action) error

	AuthorizeFnCalls atomic.Int64
}

func (m *MockAuthorizer) Authorize(role string, resource rbac.Resource, action rbac.Action) error {
	m.AuthorizeFnCalls.Add(1)
	if m.AuthorizeFn == nil {
		panic("AuthorizeFn is nil")
	}
	return m.AuthorizeFn(role, resource, action)
}

var _ domain.Authorizer = (*MockAuthorizer)(nil)
