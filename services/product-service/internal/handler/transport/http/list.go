package http

import (
	"context"

	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
)

func (h *Handler) GetProducts(
	ctx context.Context,
	req api.GetProductsRequestObject,
) (api.GetProductsResponseObject, error) {

	products, err := h.service.List(ctx, toListFilter(req.Params))
	if err != nil {
		return mapListError(h.log, err), nil
	}

	return api.GetProducts200JSONResponse(
		toProductList(products),
	), nil
}