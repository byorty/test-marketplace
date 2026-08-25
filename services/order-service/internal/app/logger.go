package app

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	return logger
}

var LoggerModule = fx.Provide(NewLogger)