package domain

import (
	"context"

	"github.com/google/uuid"
)

type MockProductRepository struct {
	
	CreateFunc func(context.Context, *Product) error
	GetByIDFunc func(context.Context, uuid.UUID) (*Product, error)
	UpdateFunc func(context.Context, *Product) (*Product, error)
	DeleteFunc func(context.Context, uuid.UUID) error
	ListFunc func(context.Context, ListFilter) (*ProductList, error)

	CreateCalls int
	GetByIDCalls int
	UpdateCalls int
	DeleteCalls int
	ListCalls int
}

func (m *MockProductRepository) Create(ctx context.Context, product *Product) error {
	m.CreateCalls++

	if m.CreateFunc == nil {
		panic("CreateFunc is nil")
	}

	return m.CreateFunc(ctx, product)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	m.GetByIDCalls++

	if m.GetByIDFunc == nil {
		panic("GetByIDFunc is nil")
	}

	return m.GetByIDFunc(ctx, id)
}

func (m *MockProductRepository) Update(ctx context.Context, product *Product) (*Product, error) {
	m.UpdateCalls++

	if m.UpdateFunc == nil {
		panic("UpdateFunc is nil")
	}

	return m.UpdateFunc(ctx, product)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.DeleteCalls++

	if m.DeleteFunc == nil {
		panic("DeleteFunc is nil")
	}

	return m.DeleteFunc(ctx, id)
}

func (m *MockProductRepository) List(ctx context.Context, filter ListFilter) (*ProductList, error) {
	m.ListCalls++

	if m.ListFunc == nil {
		panic("ListFunc is nil")
	}

	return m.ListFunc(ctx, filter)
}


