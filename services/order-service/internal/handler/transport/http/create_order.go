package http

import (
	"context"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
)

func (h *Handler) CreateOrder(
	ctx context.Context,
	req api.CreateOrderRequestObject,
) (api.CreateOrderResponseObject, error) {

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.CreateOrder401JSONResponse(
			errorResponse("unauthorized", "missing jwt claims"),
		), nil
	}

	order, err := h.service.CreateOrder(ctx, claims.UserID); 
	if err != nil {
		return mapCreateOrderError(h.log, err), nil
	}

	return api.CreateOrder201JSONResponse(
		toCreateOrderResp(order),
	), nil
}