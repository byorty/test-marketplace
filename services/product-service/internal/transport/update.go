package transport

import (
	"context"

	api "github.com/byorty/test-marketplace/services/product-service/internal/generated/openapi"
)

func (h *ProductHandler) UpdateProduct(
	ctx context.Context,
	req api.UpdateProductRequestObject,
) (api.UpdateProductResponseObject, error) {

	product, err := h.service.Update(ctx, req.Id, req.Body)
	if err != nil {
		return mapUpdateError(h.log, err), nil
	}

	return api.UpdateProduct200JSONResponse{
		Name:          product.Name,
		Description:   product.Description,
		Category:      product.Category,
		Price:         product.Price,
		DeliveryDays:  product.DeliveryDays,
		UpdatedAt:     product.UpdatedAt,
	}, nil
}