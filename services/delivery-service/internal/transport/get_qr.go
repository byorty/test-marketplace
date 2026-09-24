package transport

import (
	"context"

	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/common/rbac"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
)

func (h *DeliveryHandler) GetDeliveryQR(ctx context.Context, req api.GetDeliveryQRRequestObject) (api.GetDeliveryQRResponseObject, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.GetDeliveryQR401JSONResponse(errorResponse("unauthorized", "missing jwt claims")), nil
	}

	if err := h.authorizer.Authorize(claims.Role, rbac.ResourceDelivery, rbac.ActionGetQR); err != nil {
		return api.GetDeliveryQR403JSONResponse(errorResponse("forbidden", err.Error())), nil
	}

	qrResp, err := h.service.GetQR(ctx, claims.UserID, req.Id, claims.Role)
	if err != nil {
		return mapGetDeliveryQRError(h.log, err), nil
	}

	return api.GetDeliveryQR200JSONResponse{
		Token:     qrResp.Token,
		ExpiresAt: qrResp.ExpiresAt,
	}, nil
}
