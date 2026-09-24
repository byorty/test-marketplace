package app

import (
	orderclient "github.com/byorty/test-marketplace/services/common/client/order"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/client"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/config"
	"go.uber.org/fx"
)

func NewOrderClient(cfg *config.Config) client.Client {
	return orderclient.NewOrderClient(cfg.OrderService.URL)
}

var ClientModule = fx.Provide(NewOrderClient)
