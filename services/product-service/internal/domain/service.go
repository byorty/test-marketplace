package domain

import (
	"context"

	api "github.com/byorty/test-marketplace/services/product-service/internal/generated/openapi"
	"github.com/google/uuid"
)

type ProductService interface {
	Create(ctx context.Context, input *api.ProductCreateRequest) (*Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Update(ctx context.Context, id uuid.UUID, input *api.ProductUpdateRequest) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) (*ProductList, error)
}