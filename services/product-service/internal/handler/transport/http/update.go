package http

import (
	"context"

	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
)

func (h *Handler) UpdateProduct(
	ctx context.Context,
	req api.UpdateProductRequestObject,
) (api.UpdateProductResponseObject, error) {

	err := h.service.Update(ctx, req.Id, toUpdateInput(*req.Body))
	if err != nil {
		return mapUpdateError(h.log, err), nil
	}

	return api.UpdateProduct200JSONResponse{}, nil
}