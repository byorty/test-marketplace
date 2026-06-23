package service

import (
	"context"

	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, input *CreateProduct) (*domain.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	Update(ctx context.Context, id uuid.UUID, updates *UpdateProduct) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error)
}
