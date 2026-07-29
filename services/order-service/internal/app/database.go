package app

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/config"
	"github.com/byorty/test-marketplace/services/order-service/internal/database"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	return database.New(cfg.Postgres)
}

var DatabaseModule = fx.Provide(NewDB)