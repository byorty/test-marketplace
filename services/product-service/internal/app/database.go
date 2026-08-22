package app

import (
	"github.com/byorty/test-marketplace/services/common/database"
	"github.com/byorty/test-marketplace/services/product-service/internal/config"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

func NewDB(cfg *config.Config) (*bun.DB, error) {
	return database.New(database.PostgresConfig(cfg.Postgres))
}

var DatabaseModule = fx.Provide(NewDB)