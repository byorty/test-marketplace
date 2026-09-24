package transport

import (
	"context"

	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/common/rbac"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
)

func (h *DeliveryHandler) VerifyDeliveryQR(ctx context.Context, req api.VerifyDeliveryQRRequestObject) (api.VerifyDeliveryQRResponseObject, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.VerifyDeliveryQR401JSONResponse(errorResponse("unauthorized", "missing jwt claims")), nil
	}

	if err := h.authorizer.Authorize(claims.Role, rbac.ResourceDelivery, rbac.ActionVerifyQR); err != nil {
		return api.VerifyDeliveryQR403JSONResponse(errorResponse("forbidden", err.Error())), nil
	}

	delivery, err := h.service.VerifyQR(ctx, req.Body.Token)
	if err != nil {
		return mapVerifyDeliveryQRError(h.log, err), nil
	}

	return api.VerifyDeliveryQR200JSONResponse{
		DeliveryId: delivery.ID,
		OrderId:    delivery.OrderID,
		Verified:   true,
	}, nil
}
