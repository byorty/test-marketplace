package http

import (
	"context"

	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	"github.com/byorty/test-marketplace/services/product-service/internal/service"
	"github.com/google/uuid"
)

type MockService struct {
	CreateFunc func(context.Context, *service.CreateProduct) (*domain.Product, error)
	GetByIDFunc func(context.Context, uuid.UUID) (*domain.Product, error)
	UpdateFunc func(context.Context, uuid.UUID, *service.UpdateProduct) error
	DeleteFunc func(context.Context, uuid.UUID) error
	ListFunc func(context.Context, domain.ListFilter) (*domain.ProductList, error)

	CreateCalls int
	GetByIDCalls int
	UpdateCalls int
	DeleteCalls int
	ListCalls int
}

func (m *MockService) Create(ctx context.Context, input *service.CreateProduct) (*domain.Product, error) {
	m.CreateCalls++

	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}

	return nil, nil
}

func (m *MockService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	m.GetByIDCalls++

	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}

	return nil, nil
}

func (m *MockService) Update(ctx context.Context, id uuid.UUID, input *service.UpdateProduct) error {
	m.UpdateCalls++

	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, input)
	}

	return nil
}

func (m *MockService) Delete(ctx context.Context, id uuid.UUID) error {
	m.DeleteCalls++

	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}

	return nil
}

func (m *MockService) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
	m.ListCalls++

	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}

	return nil, nil
}