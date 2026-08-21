package domain

import (
	"context"

	"github.com/google/uuid"
)

type OrderRepository interface {
	//Cart
	AddToCart(ctx context.Context, userID uuid.UUID, item *CartItem) error
	GetCart(ctx context.Context, userID uuid.UUID) ([]CartItem, error)
	GetCartItem(ctx context.Context, userID, productID uuid.UUID) (*CartItem, error)
	RemoveFromCart(ctx context.Context, userID uuid.UUID, cartItemID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
	//Orders
	CreateOrder(ctx context.Context, order *Order) error
	CreateOrderItems(ctx context.Context, items []OrderItem) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error)

	Transaction(ctx context.Context, fn func(repo OrderRepository) error) error
}


