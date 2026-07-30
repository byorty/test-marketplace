package service

import (
	"context"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain/order"
	"github.com/google/uuid"
)

type MockRepository struct {
	
	AddToCartFn func(ctx context.Context, item *order.CartItem) error
	GetCartFn func(ctx context.Context, userID uuid.UUID) ([]order.CartItem, error)
	RemoveFromCartFn func(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error
	ClearCartFn func(ctx context.Context, userID uuid.UUID) error

	CreateOrderFn func(ctx context.Context, order *order.Order) error
	CreateOrderItemsFn func(ctx context.Context, items []order.OrderItem) error
	GetOrderByIDFn func(ctx context.Context, id uuid.UUID) (*order.Order, error)
	GetOrderItemsFn func(ctx context.Context, orderID uuid.UUID) ([]order.OrderItem, error)

	TransactionFn func(ctx context.Context, fn func(repo order.Repository) error) error

	AddToCartFnCalls int
	GetCartFnCalls int
	RemoveFromCartFnCalls int
	ClearCartFnCalls int

	CreateOrderFnCalls int
	CreateOrderItemsFnCalls int
	GetOrderByIDFnCalls int
	GetOrderItemsFnCalls int

	TransactionFnCalls int

	Self order.Repository
}

func (m *MockRepository) AddToCart(ctx context.Context, item *order.CartItem) error {
	m.AddToCartFnCalls++

	if m.AddToCartFn == nil {
		panic("AddToCartF is nil")
	}

	return m.AddToCartFn(ctx, item)
}

func (m *MockRepository) GetCart(ctx context.Context, userID uuid.UUID) ([]order.CartItem, error) {
	m.GetCartFnCalls++

	if m.GetCartFn == nil {
		panic("GetCartF is nil")
	}

	return m.GetCartFn(ctx, userID)
}

func (m *MockRepository) RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error {
	m.RemoveFromCartFnCalls++

	if m.RemoveFromCartFn == nil {
		panic("RemoveFromCartF is nil")
	}

	return m.RemoveFromCartFn(ctx, userID, productID)
}

func (m *MockRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	m.ClearCartFnCalls++

	if m.ClearCartFn == nil {
		panic("ClearCartF is nil")
	}

	return m.ClearCartFn(ctx, userID)
}

func (m *MockRepository) CreateOrder(ctx context.Context, order *order.Order) error {
	m.CreateOrderFnCalls++

	if m.CreateOrderFn == nil {
		panic("CreateOrderF is nil")
	}

	return m.CreateOrderFn(ctx, order)
}

func (m *MockRepository) CreateOrderItems(ctx context.Context, items []order.OrderItem) error {
	m.CreateOrderItemsFnCalls++

	if m.CreateOrderItemsFn == nil {
		panic("CreateOrderItemsF is nil")
	}

	return m.CreateOrderItemsFn(ctx, items)
}

func (m *MockRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	m.GetOrderByIDFnCalls++

	if m.GetOrderByIDFn == nil {
		panic("GetOrderByIDF is nil")
	}

	return m.GetOrderByIDFn(ctx, id)
}

func (m *MockRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]order.OrderItem, error) {
	m.GetOrderItemsFnCalls++

	if m.GetOrderItemsFn == nil {
		panic("GetOrderItemsF is nil")
	}

	return m.GetOrderItemsFn(ctx, orderID)
}

func (m *MockRepository) Transaction(ctx context.Context, fn func(repo order.Repository) error) error {
	m.TransactionFnCalls++

	if m.TransactionFn == nil {
		panic("TransactionFn is nil")
	}

	if m.Self == nil {
		m.Self = m
	}

	return m.TransactionFn(ctx, func(order.Repository) error {
		return fn(m.Self)
	})
}
