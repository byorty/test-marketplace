package transport

import (
	"context"

	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
)

func (h *DeliveryHandler) UpdateDeliveryStatus(ctx context.Context, req api.UpdateDeliveryStatusRequestObject) (api.UpdateDeliveryStatusResponseObject, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.UpdateDeliveryStatus401JSONResponse(errorResponse("unauthorized", "missing jwt claims")), nil
	}

	if err := h.authorizer.Authorize(claims.Role, rbac.ResourceDelivery, rbac.ActionUpdateStatus); err != nil {
		return api.UpdateDeliveryStatus403JSONResponse(errorResponse("forbidden", err.Error())), nil
	}

	status := domain.DeliveryStatus(req.Body.Status)

	delivery, err := h.service.UpdateStatus(ctx, claims.UserID, req.Id, claims.Role, status)
	if err != nil {
		return mapUpdateDeliveryStatusError(h.log, err), nil
	}

	return api.UpdateDeliveryStatus200JSONResponse(toDeliveryResponse(delivery)), nil
}
