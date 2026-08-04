package app

import (
	httptransport "github.com/byorty/test-marketplace/services/product-service/internal/transport"
	"go.uber.org/fx"
)

var HandlerModule = fx.Provide(httptransport.New)