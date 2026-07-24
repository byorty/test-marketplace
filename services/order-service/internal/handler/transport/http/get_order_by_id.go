package http

import (
	"context"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
)

func (h *Handler) GetOrderByID(
	ctx context.Context,
	req api.GetOrderByIDRequestObject,
) (api.GetOrderByIDResponseObject, error) {

	_, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.GetOrderByID401JSONResponse(
			errorResponse("unauthorized", "missing jwt claims"),
		), nil
	}

	order, err := h.service.GetOrderByID(ctx, req.Id)
	if err != nil {
		return mapGetOrderByIDError(h.log, err), nil
	}

	return api.GetOrderByID200JSONResponse(
		toOrderResponse(order),
	), nil
}