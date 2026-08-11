package domain

import (
	"context"

	"github.com/google/uuid"
)

type OrderService interface {
	AddToCart(ctx context.Context, userID uuid.UUID, item *CartItem) error
	GetCart(ctx context.Context, userID uuid.UUID) (*Cart, error)
	RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error

	GetOrderByID(ctx context.Context, userID, orderID uuid.UUID) (*Order, error)
	CreateOrder(ctx context.Context, userID uuid.UUID) (*Order, error)
}