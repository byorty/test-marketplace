package app

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/byorty/test-marketplace/services/order-service/internal/service"
	"go.uber.org/fx"
)
var ServiceModule = fx.Provide(
	fx.Annotate(
		service.New,
		fx.As(new(domain.OrderService)),
	),
)