package app

import (
	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	"github.com/byorty/test-marketplace/services/product-service/internal/repository"
	"go.uber.org/fx"
)

var RepositoryModule = fx.Provide(
	fx.Annotate(
		repository.New,
		fx.As(new(domain.ProductRepository)),
	),
)