package app

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/config"
	"go.uber.org/fx"
)

var ConfigModule = fx.Provide(config.MustLoad)