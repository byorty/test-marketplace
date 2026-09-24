package transport

import (
	"context"

	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
)

func (h *DeliveryHandler) GetDeliveries(ctx context.Context, req api.GetDeliveriesRequestObject) (api.GetDeliveriesResponseObject, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return api.GetDeliveries401JSONResponse(errorResponse("unauthorized", "missing jwt claims")), nil
	}

	if err := h.authorizer.Authorize(claims.Role, rbac.ResourceDelivery, rbac.ActionView); err != nil {
		return api.GetDeliveries403JSONResponse(errorResponse("forbidden", err.Error())), nil
	}

	filter := domain.DeliveryListFilter{
		Page:     1,
		PageSize: 20,
	}

	if req.Params.Page != nil {
		filter.Page = *req.Params.Page
	}
	if req.Params.PageSize != nil {
		filter.PageSize = *req.Params.PageSize
	}
	if req.Params.Status != nil {
		s := domain.DeliveryStatus(*req.Params.Status)
		filter.Status = &s
	}
	if req.Params.IsLate != nil {
		filter.IsLate = req.Params.IsLate
	}
	if req.Params.OrderId != nil {
		filter.OrderID = req.Params.OrderId
	}

	result, err := h.service.List(ctx, claims.UserID, claims.Role, filter)
	if err != nil {
		return mapGetDeliveriesError(h.log, err), nil
	}

	items := make([]api.DeliveryResponse, 0, len(result.Items))
	for _, d := range result.Items {
		items = append(items, toDeliveryResponse(d))
	}

	return api.GetDeliveries200JSONResponse{
		Items:    items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}
