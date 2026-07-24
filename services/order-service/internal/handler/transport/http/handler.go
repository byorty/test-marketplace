package http

import (
	"context"
	"log/slog"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain/order"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
	"github.com/google/uuid"
)

type OrderService interface {
	AddToCart(ctx context.Context, item *order.CartItem) error
	GetCart(ctx context.Context, userID uuid.UUID) (*order.Cart, error)
	RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error

	GetOrderByID(ctx context.Context, id uuid.UUID) (*order.Order, error)
	CreateOrder(ctx context.Context, userID uuid.UUID) (*order.Order, error)
}

type Handler struct {
	service OrderService
	log *slog.Logger
}

func New(service OrderService, log *slog.Logger) api.StrictServerInterface {
	return &Handler{
		service: service,
		log: log,
	}
}