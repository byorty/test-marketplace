package app

import (
	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"

	"go.uber.org/fx"
)

func NewDomainAuthorizer(
    authorizer *rbac.Authorizer,
) domain.Authorizer {
    return authorizer
}

var RBACModule = fx.Provide(
    rbac.NewEnforcer,
    rbac.New,
    NewDomainAuthorizer,
)