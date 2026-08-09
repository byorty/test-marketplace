package app

import (
	"github.com/byorty/test-marketplace/services/common/rbac"
	"go.uber.org/fx"
)

var RBACModule = fx.Options(

	fx.Provide(rbac.NewEnforcer),

	fx.Provide(rbac.New),
)