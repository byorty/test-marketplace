package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/byorty/test-marketplace/services/product-service/internal/config"
)

func New(cfg config.LogConfig) (*slog.Logger, error) {
	var level slog.Level

	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug

	case "info":
		level = slog.LevelInfo

	case "warn":
		level = slog.LevelWarn

	case "error":
		level = slog.LevelError

	default:
		return nil, fmt.Errorf("unknown log level %q", cfg.Level)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})

	return slog.New(handler), nil
}