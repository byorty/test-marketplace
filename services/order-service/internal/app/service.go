package app

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/service"
	"go.uber.org/fx"
)

var ServiceModule = fx.Provide(service.New)