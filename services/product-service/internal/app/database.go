package app

import (
	"github.com/byorty/test-marketplace/services/product-service/internal/config"
	"github.com/byorty/test-marketplace/services/product-service/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func NewDB(cfg *config.Config) (*pgxpool.Pool, error) {
	return database.New(cfg.Postgres)
}

var DatabaseModule = fx.Provide(NewDB)