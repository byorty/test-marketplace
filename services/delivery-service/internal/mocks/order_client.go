package mocks

import (
	"context"
	"sync/atomic"

	orderclient "github.com/byorty/test-marketplace/services/common/client/order"
	"github.com/google/uuid"
)

type MockOrderClient struct {
	GetOrderByIDFn      func(ctx context.Context, orderID uuid.UUID) (*orderclient.OrderResponse, error)
	UpdateOrderStatusFn func(ctx context.Context, orderID uuid.UUID, status string) error

	GetOrderByIDFnCalls      atomic.Int64
	UpdateOrderStatusFnCalls atomic.Int64
}

func (m *MockOrderClient) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*orderclient.OrderResponse, error) {
	m.GetOrderByIDFnCalls.Add(1)
	if m.GetOrderByIDFn == nil {
		panic("GetOrderByIDFn is nil")
	}
	return m.GetOrderByIDFn(ctx, orderID)
}

func (m *MockOrderClient) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	m.UpdateOrderStatusFnCalls.Add(1)
	if m.UpdateOrderStatusFn == nil {
		panic("UpdateOrderStatusFn is nil")
	}
	return m.UpdateOrderStatusFn(ctx, orderID, status)
}
