package http

import (
	"context"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
)

func (h *Handler) RemoveFromCart(
	ctx context.Context, 
	req api.RemoveFromCartRequestObject,
) (api.RemoveFromCartResponseObject, error) {

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.RemoveFromCart401JSONResponse(
			errorResponse("unauthorized", "missing jwt claims"),
		), nil
	}

	if err := h.service.RemoveFromCart(ctx, claims.UserID, req.ProductId); err != nil {
		return mapRemoveFromCartError(h.log, err), nil
	}

	return api.RemoveFromCart204Response{}, nil
}