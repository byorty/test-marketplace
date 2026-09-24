package app

import (
	httptransport "github.com/byorty/test-marketplace/services/delivery-service/internal/transport"
	"go.uber.org/fx"
)

var HandlerModule = fx.Provide(httptransport.NewDeliveryHandler)
