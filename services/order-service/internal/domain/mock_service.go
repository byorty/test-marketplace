package domain

import (
	"context"

	"github.com/google/uuid"
)

type MockOrderService struct {
	AddToCartFunc      func(ctx context.Context, item *CartItem) error
	GetCartFunc        func(ctx context.Context, userID uuid.UUID) (*Cart, error)
	RemoveFromCartFunc func(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error
	GetOrderByIDFunc   func(ctx context.Context, id uuid.UUID) (*Order, error)
	CreateOrderFunc    func(ctx context.Context, userID uuid.UUID) (*Order, error)

	AddToCartCalls      int
	GetCartCalls        int
	RemoveFromCartCalls int
	GetOrderByIDCalls   int
	CreateOrderCalls    int
}

func (m *MockOrderService) AddToCart(ctx context.Context, item *CartItem) error {
	m.AddToCartCalls++

	if m.AddToCartFunc != nil {
		return m.AddToCartFunc(ctx, item)
	}

	return nil
}

func (m *MockOrderService) GetCart(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	m.GetCartCalls++

	if m.GetCartFunc != nil {
		return m.GetCartFunc(ctx, userID)
	}

	return nil, nil
}

func (m *MockOrderService) RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error {
	m.RemoveFromCartCalls++

	if m.RemoveFromCartFunc != nil {
		return m.RemoveFromCartFunc(ctx, userID, productID)
	}

	return nil
}

func (m *MockOrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	m.GetOrderByIDCalls++

	if m.GetOrderByIDFunc != nil {
		return m.GetOrderByIDFunc(ctx, id)
	}

	return nil, nil
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID uuid.UUID) (*Order, error) {
	m.CreateOrderCalls++

	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, userID)
	}

	return nil, nil
}