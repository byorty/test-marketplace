package product

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Update(ctx context.Context, product *Product, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) (*ProductList, error)
}
