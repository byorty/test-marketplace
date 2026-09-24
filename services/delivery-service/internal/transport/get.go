package transport

import (
	"context"

	"github.com/byorty/test-marketplace/services/common/auth"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
)

func (h *DeliveryHandler) GetDelivery(ctx context.Context, req api.GetDeliveryRequestObject) (api.GetDeliveryResponseObject, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.GetDelivery401JSONResponse(errorResponse("unauthorized", "missing jwt claims")), nil
	}

	delivery, err := h.service.GetByID(ctx, claims.UserID, req.Id, claims.Role)
	if err != nil {
		return mapGetDeliveryError(h.log, err), nil
	}

	return api.GetDelivery200JSONResponse(toDeliveryResponse(delivery)), nil
}
