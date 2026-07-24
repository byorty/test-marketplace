package http

import (
	"context"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
)

func (h *Handler) GetCart(
	ctx context.Context,
	req api.GetCartRequestObject,
) (api.GetCartResponseObject, error) {

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.GetCart401JSONResponse(
			errorResponse("unauthorized", "missing jwt claims"),
			), nil
	}
	
	cart, err := h.service.GetCart(ctx, claims.UserID)
	if err != nil {
		return mapGetCartError(h.log, err), nil
	}

	

	return api.GetCart200JSONResponse(
		toCartResponse(cart),
	), nil
}