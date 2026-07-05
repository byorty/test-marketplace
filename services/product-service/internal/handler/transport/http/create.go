package http

import (
	"context"

	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
)

func (h *Handler) CreateProduct(
	ctx context.Context, 
	req api.CreateProductRequestObject,
	) (api.CreateProductResponseObject, error) {

	product, err := h.service.Create(ctx, toCreateInput(*req.Body))
	if err != nil {
		return mapCreateError(h.log, err), nil
	}

	return api.CreateProduct201JSONResponse(
		toResponse(product),
	), nil
}