package transport

import (
	"context"

	"github.com/byorty/test-marketplace/services/common/auth"
	rbac "github.com/byorty/test-marketplace/services/common/rbac"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
)

func (h *OrderHandler) GetOrderByID(
    ctx context.Context,
    req api.GetOrderByIDRequestObject,
) (api.GetOrderByIDResponseObject, error) {

    claims, ok := auth.ClaimsFromContext(ctx)
    if !ok {
        return api.GetOrderByID401JSONResponse(
            errorResponse("unauthorized", "missing jwt claims"),
        ), nil
    }

    if err := h.authorizer.Authorize(
        claims.Role,
        rbac.ResourceOrder,
        rbac.ActionView,
    ); err != nil {
        return api.GetOrderByID403JSONResponse(
            errorResponse("forbidden", err.Error()),
        ), nil
    }

    order, err := h.service.GetOrderByID(ctx, claims.UserID, req.Id)
    if err != nil {
        return mapGetOrderByIDError(h.log, err), nil
    }

    if claims.Role != "employee" && order.UserID != claims.UserID {
        return api.GetOrderByID403JSONResponse(
            errorResponse("forbidden", "order does not belong to user"),
        ), nil
    }

    return api.GetOrderByID200JSONResponse(
        toOrderResponse(order),
    ), nil
}