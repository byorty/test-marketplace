package http

import (
	"context"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
)

func (h *Handler) AddToCart(
	ctx context.Context, 
	req api.AddToCartRequestObject,
	) (api.AddToCartResponseObject, error) {

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.AddToCart401JSONResponse(
			errorResponse("unauthorized", "missing jwt claims"),
			), nil
	}

	input := toCartItemInput(
		claims.UserID,
		*req.Body,
	)
	
	if err := h.service.AddToCart(ctx, input); err != nil {
		return mapAddToCartError(h.log, err), nil
	}

	return api.AddToCart201Response{}, nil
}