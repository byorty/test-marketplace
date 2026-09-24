package transport

import (
	"context"

	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
)

func (h *DeliveryHandler) CreateDelivery(ctx context.Context, req api.CreateDeliveryRequestObject) (api.CreateDeliveryResponseObject, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.CreateDelivery401JSONResponse(errorResponse("unauthorized", "missing jwt claims")), nil
	}

	input := &domain.DeliveryCreateInput{
		UserID:                claims.UserID,
		OrderID:               req.Body.OrderId,
		PickupAddress:         req.Body.PickupAddress,
		EstimatedDeliveryDate: req.Body.EstimatedDeliveryDate,
		Role:                  claims.Role,
	}

	delivery, err := h.service.Create(ctx, input)
	if err != nil {
		return mapCreateDeliveryError(h.log, err), nil
	}

	return api.CreateDelivery201JSONResponse(toDeliveryResponse(delivery)), nil
}
