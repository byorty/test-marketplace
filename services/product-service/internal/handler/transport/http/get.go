package http

import (
	"context"

	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
)

func (h *Handler) GetByID(
	ctx context.Context,
	req api.GetByIDRequestObject,
) (api.GetByIDResponseObject, error) {

	product, err := h.service.GetByID(ctx, req.Id)
	if err != nil {
		return mapGetError(h.log, err), nil
	}

	return api.GetByID200JSONResponse(
		toResponse(product),
	), nil
}