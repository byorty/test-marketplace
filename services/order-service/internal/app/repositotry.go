package app

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/repository/postgres"
	"go.uber.org/fx"
)

var RepositoryModule = fx.Provide(postgres.New)