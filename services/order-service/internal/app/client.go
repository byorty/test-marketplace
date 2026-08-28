package app

import (
	"github.com/byorty/test-marketplace/services/common/client/product"
	"github.com/byorty/test-marketplace/services/order-service/internal/client"
	"go.uber.org/fx"
)

var ClientModule = fx.Provide(
	fx.Annotate(
		product.NewProductClient,
		fx.As(new(client.Client)),
	),
)