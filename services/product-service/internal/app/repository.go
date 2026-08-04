package app

import (
	"github.com/byorty/test-marketplace/services/product-service/internal/repository"
	"go.uber.org/fx"
)

var RepositoryModule = fx.Provide(repository.New)