package http

import (
	"log/slog"

	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	s "github.com/byorty/test-marketplace/services/product-service/internal/service"
)

type Handler struct {
	service s.Service
	log     *slog.Logger
}

func New(service s.Service, log *slog.Logger) api.StrictServerInterface {
	return &Handler{
		service: service,
		log:     log,
	}
}
