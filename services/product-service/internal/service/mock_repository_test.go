package service

import (
	"context"

	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	"github.com/google/uuid"
)

type MockRepository struct {
	CreateFunc func(context.Context, *domain.Product) error
	GetByIDFunc func(context.Context, uuid.UUID) (*domain.Product, error)
	UpdateFunc func(context.Context, *domain.Product) error
	DeleteFunc func(context.Context, uuid.UUID) error
	ListFunc func(context.Context, domain.ListFilter) (*domain.ProductList, error)

	CreateCalls int
	GetByIDCalls int
	UpdateCalls int
	DeleteCalls int
	ListCalls int
}

func (m *MockRepository) Create(ctx context.Context, product *domain.Product) error {
	m.CreateCalls++

	if m.CreateFunc == nil {
		panic("CreateFunc is nil")
	}

	return m.CreateFunc(ctx, product)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	m.GetByIDCalls++

	if m.GetByIDFunc == nil {
		panic("GetByIDFunc is nil")
	}

	return m.GetByIDFunc(ctx, id)
}

func (m *MockRepository) Update(ctx context.Context, product *domain.Product) error {
	m.UpdateCalls++

	if m.UpdateFunc == nil {
		panic("UpdateFunc is nil")
	}

	return m.UpdateFunc(ctx, product)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.DeleteCalls++

	if m.DeleteFunc == nil {
		panic("DeleteFunc is nil")
	}

	return m.DeleteFunc(ctx, id)
}

func (m *MockRepository) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
	m.ListCalls++

	if m.ListFunc == nil {
		panic("ListFunc is nil")
	}

	return m.ListFunc(ctx, filter)
}
