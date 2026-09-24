package transport

import (
	"context"

	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
)

func (h *DeliveryHandler) RescheduleDelivery(ctx context.Context, req api.RescheduleDeliveryRequestObject) (api.RescheduleDeliveryResponseObject, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.RescheduleDelivery401JSONResponse(errorResponse("unauthorized", "missing jwt claims")), nil
	}

	if err := h.authorizer.Authorize(claims.Role, rbac.ResourceDelivery, rbac.ActionReschedule); err != nil {
		return api.RescheduleDelivery403JSONResponse(errorResponse("forbidden", err.Error())), nil
	}

	input := &domain.RescheduleInput{
		NewDate: req.Body.NewDate,
	}
	if req.Body.Reason != nil {
		input.Reason = *req.Body.Reason
	}

	delivery, err := h.service.Reschedule(ctx, claims.UserID, req.Id, claims.Role, input)
	if err != nil {
		return mapRescheduleDeliveryError(h.log, err), nil
	}

	return api.RescheduleDelivery200JSONResponse(toDeliveryResponse(delivery)), nil
}
