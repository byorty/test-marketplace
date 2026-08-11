package transport

import (
	"context"

	api "github.com/byorty/test-marketplace/services/product-service/internal/generated/openapi"
)

func (h *ProductHandler) DeleteProduct(
	ctx context.Context,
	req api.DeleteProductRequestObject,
) (api.DeleteProductResponseObject, error) {

	if err := h.service.Delete(ctx, req.Id); err != nil {
		return mapDeleteError(h.log, err), nil
	}

	return api.DeleteProduct204Response{}, nil
}