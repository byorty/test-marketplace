package app

import (
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/repository"
	"go.uber.org/fx"
)

var RepositoryModule = fx.Provide(
	fx.Annotate(
		repository.New,
		fx.As(new(domain.DeliveryRepository)),
	),
)
