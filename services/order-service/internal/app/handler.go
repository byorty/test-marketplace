package app

import (
	httptransport "github.com/byorty/test-marketplace/services/order-service/internal/handler/transport/http"
	"go.uber.org/fx"
)

var HandlerModule = fx.Provide(httptransport.New)