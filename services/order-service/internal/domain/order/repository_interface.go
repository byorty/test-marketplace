package order

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	//Cart
	AddToCart(ctx context.Context, items *CartItem) error
	GetCart(ctx context.Context, userID uuid.UUID) ([]CartItem, error)
	RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
	//Orders
	CreateOrder(ctx context.Context, order *Order) error
	CreateOrderItems(ctx context.Context, items []OrderItem) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error)
}