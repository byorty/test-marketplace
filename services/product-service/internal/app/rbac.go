package app

import (
	"github.com/byorty/test-marketplace/services/product-service/internal/rbac"
	"go.uber.org/fx"
)

var RBACModule = fx.Options(

	fx.Provide(rbac.NewEnforcer),

	fx.Provide(rbac.New),
)