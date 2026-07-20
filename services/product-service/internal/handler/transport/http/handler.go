package http

import (
	"context"
	"log/slog"

	"github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	"github.com/byorty/test-marketplace/services/product-service/internal/service"
	s "github.com/byorty/test-marketplace/services/product-service/internal/service"
	"github.com/google/uuid"
)

type ProductService interface {
	Create(ctx context.Context, input *service.CreateProduct) (*product.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error)
	Update(ctx context.Context, id uuid.UUID, input *service.UpdateProduct) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter product.ListFilter) (*product.ProductList, error)
}

type Handler struct {
	service ProductService
	log     *slog.Logger
}

func New(service *s.Service, log *slog.Logger) api.StrictServerInterface {
	return &Handler{
		service: service,
		log:     log,
	}
}
