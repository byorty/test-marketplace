package app

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	"github.com/byorty/test-marketplace/services/order-service/internal/config"
	"go.uber.org/fx"
)

func NewJWTValidator(cfg *config.Config) *auth.Validator {
	return auth.NewValidator(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
	)
}

var AuthModule = fx.Provide(NewJWTValidator)
