package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (a *App) Run() error {
	const op = "App.Run"

	go func() {
		a.log.Info("http server started", slog.String("addr", a.server.Addr))

		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("http server stopped", slog.Any("error", err))
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(), 
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	a.log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	) 
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("%s: shutdown server: %w", op, err)
	}

	if err := a.sqlDB.Close(); err != nil {
		a.log.Error("close database", slog.Any("error", err))
	}

	a.log.Info("http server stopped")

	return nil
}