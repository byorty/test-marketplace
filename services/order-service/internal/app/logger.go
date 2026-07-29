package app

import (
	"log/slog"
	"os"

	"go.uber.org/fx"
)

func NewLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)
}

var LoggerModule = fx.Provide(NewLogger)