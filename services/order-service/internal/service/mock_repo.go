package service

import (
	"context"

	client "github.com/byorty/test-marketplace/services/common/client/product/generated"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/google/uuid"
)

type MockOrderRepository struct {
	
	AddToCartFn func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error
	GetCartFn func(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error)
	GetCartItemFn func(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*domain.CartItem, error)
	RemoveFromCartFn func(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error
	ClearCartFn func(ctx context.Context, userID uuid.UUID) error

	CreateOrderFn func(ctx context.Context, order *domain.Order) error
	CreateOrderItemsFn func(ctx context.Context, items []domain.OrderItem) error
	GetOrderByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetOrderItemsFn func(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error)

	TransactionFn func(ctx context.Context, fn func(repo domain.OrderRepository) error) error

	AddToCartFnCalls int
	GetCartFnCalls int
	GetCartItemFnCalls int
	RemoveFromCartFnCalls int
	ClearCartFnCalls int

	CreateOrderFnCalls int
	CreateOrderItemsFnCalls int
	GetOrderByIDFnCalls int
	GetOrderItemsFnCalls int

	TransactionFnCalls int

	Self domain.OrderRepository
}


func (m *MockOrderRepository) AddToCart(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
	m.AddToCartFnCalls++

	if m.AddToCartFn == nil {
		panic("AddToCartF is nil")
	}

	return m.AddToCartFn(ctx, userID, item)
}

func (m *MockOrderRepository) GetCart(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error) {
	m.GetCartFnCalls++

	if m.GetCartFn == nil {
		panic("GetCartF is nil")
	}

	return m.GetCartFn(ctx, userID)
}

func (m *MockOrderRepository) GetCartItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*domain.CartItem, error) {
	m.GetCartItemFnCalls++

	if m.GetCartItemFn == nil {
		panic("GetCartItem is nil")
	}

	return m.GetCartItemFn(ctx, userID, productID)
}

func (m *MockOrderRepository) RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error {
	m.RemoveFromCartFnCalls++

	if m.RemoveFromCartFn == nil {
		panic("RemoveFromCartF is nil")
	}

	return m.RemoveFromCartFn(ctx, userID, productID)
}

func (m *MockOrderRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	m.ClearCartFnCalls++

	if m.ClearCartFn == nil {
		panic("ClearCartF is nil")
	}

	return m.ClearCartFn(ctx, userID)
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	m.CreateOrderFnCalls++

	if m.CreateOrderFn == nil {
		panic("CreateOrderF is nil")
	}

	return m.CreateOrderFn(ctx, order)
}

func (m *MockOrderRepository) CreateOrderItems(ctx context.Context, items []domain.OrderItem) error {
	m.CreateOrderItemsFnCalls++

	if m.CreateOrderItemsFn == nil {
		panic("CreateOrderItemsF is nil")
	}

	return m.CreateOrderItemsFn(ctx, items)
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	m.GetOrderByIDFnCalls++

	if m.GetOrderByIDFn == nil {
		panic("GetOrderByIDF is nil")
	}

	return m.GetOrderByIDFn(ctx, id)
}

func (m *MockOrderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	m.GetOrderItemsFnCalls++

	if m.GetOrderItemsFn == nil {
		panic("GetOrderItemsF is nil")
	}

	return m.GetOrderItemsFn(ctx, orderID)
}

func (m *MockOrderRepository) Transaction(ctx context.Context, fn func(repo domain.OrderRepository) error) error {
	m.TransactionFnCalls++

	if m.TransactionFn == nil {
		panic("TransactionFn is nil")
	}

	if m.Self == nil {
		m.Self = m
	}

	return m.TransactionFn(ctx, func(repo domain.OrderRepository) error {
		return fn(m.Self)
	})
}

type MockProductClient struct {
	GetProductFn func(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error) 

	GetProductFnCalls int
}

func (m *MockProductClient) GetProduct(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error) {
	m.GetProductFnCalls++

	if m.GetProductFn == nil {
		panic("GetProductFn is nil")
	}

	return m.GetProductFn(ctx, id)
}
