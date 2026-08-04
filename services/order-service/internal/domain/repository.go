package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrderRepository interface {
	//Cart
	AddToCart(ctx context.Context, item *CartItem) error
	GetCart(ctx context.Context, userID uuid.UUID) ([]CartItem, error)
	RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
	//Orders
	CreateOrder(ctx context.Context, order *Order) error
	CreateOrderItems(ctx context.Context, items []OrderItem) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error)

	Transaction(ctx context.Context, fn func(repo OrderRepository) error) error
}

type DBTX interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}
