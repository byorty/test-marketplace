package app

import (
	"github.com/byorty/test-marketplace/services/delivery-service/internal/config"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/qr"
	"go.uber.org/fx"
)

func NewQRGenerator(cfg *config.Config) (*qr.Generator, error) {
	return qr.NewGenerator(cfg.QR.SecretKey, cfg.QR.TokenTTL)
}

func NewQRVerifier(cfg *config.Config) (*qr.Verifier, error) {
	return qr.NewVerifier(cfg.QR.SecretKey)
}

var QRModule = fx.Provide(NewQRGenerator, NewQRVerifier)
