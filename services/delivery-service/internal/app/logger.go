package app

import (
	"context"

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

func RegisterLoggerLifecycle(lifecycle fx.Lifecycle, logger *zap.Logger) {
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return logger.Sync()
		},
	})
}

var LoggerModule = fx.Options(
	fx.Provide(NewLogger),
	fx.Invoke(RegisterLoggerLifecycle),
)
