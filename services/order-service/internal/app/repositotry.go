package app

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/repository"
	"go.uber.org/fx"
)

var RepositoryModule = fx.Provide(repository.New)