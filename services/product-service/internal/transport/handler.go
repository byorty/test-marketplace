package transport

import (
	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated/openapi"
	s "github.com/byorty/test-marketplace/services/product-service/internal/service"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service domain.ProductService
	log     *zap.Logger
}

func New(service *s.ProductService, log *zap.Logger) api.StrictServerInterface {
	return &ProductHandler{
		service: service,
		log:     log.Named("product-handler"),
	}
}
