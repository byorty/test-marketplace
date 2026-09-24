package app

import (
	"github.com/byorty/test-marketplace/services/delivery-service/internal/config"
	"go.uber.org/fx"
)

var ConfigModule = fx.Provide(config.Load)
