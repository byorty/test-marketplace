package mocks

import (
	"context"

	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated/openapi"
	"github.com/google/uuid"
)

type MockProductService struct {
	CreateFunc func(context.Context, *api.ProductCreateRequest) (*domain.Product, error)
	GetByIDFunc func(context.Context, uuid.UUID) (*domain.Product, error)
	UpdateFunc func(context.Context, uuid.UUID, *api.ProductUpdateRequest) (*domain.Product, error)
	DeleteFunc func(context.Context, uuid.UUID) error
	ListFunc func(context.Context, domain.ListFilter) (*domain.ProductList, error)

	CreateCalls int
	GetByIDCalls int
	UpdateCalls int
	DeleteCalls int
	ListCalls int
}

func (m *MockProductService) Create(ctx context.Context, input *api.ProductCreateRequest) (*domain.Product, error) {
	m.CreateCalls++

	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}

	return nil, nil
}

func (m *MockProductService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	m.GetByIDCalls++

	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}

	return nil, nil
}

func (m *MockProductService) Update(ctx context.Context, id uuid.UUID, input *api.ProductUpdateRequest) (*domain.Product, error) {
	m.UpdateCalls++

	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, input)
	}

	return nil, nil
}

func (m *MockProductService) Delete(ctx context.Context, id uuid.UUID) error {
	m.DeleteCalls++

	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}

	return nil
}

func (m *MockProductService) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
	m.ListCalls++

	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}

	return nil, nil
}