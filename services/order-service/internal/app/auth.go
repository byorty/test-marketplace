package app

import (
	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/order-service/internal/config"
	"go.uber.org/fx"
)

func NewJWTValidator(cfg *config.Config) (*auth.Validator, error) {
	publicKey, err := auth.LoadPublicKey(cfg.JWT.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	return auth.NewValidator(
		publicKey,
		cfg.JWT.Issuer,
	), nil
}

var AuthModule = fx.Provide(NewJWTValidator)
