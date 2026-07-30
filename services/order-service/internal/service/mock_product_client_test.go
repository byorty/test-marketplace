package service

import (
	"context"

	"github.com/byorty/test-marketplace/services/order-service/internal/client/product"
	"github.com/google/uuid"
)

type MockProductClient struct {
	GetProductFn func(ctx context.Context, id uuid.UUID) (*product.Product, error) 

	GetProductFnCalls int
}

func (m *MockProductClient) GetProduct(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	m.GetProductFnCalls++

	if m.GetProductFn == nil {
		panic("GetProductFn is nil")
	}

	return m.GetProductFn(ctx, id)
}
