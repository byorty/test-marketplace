package transport

import (
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
)

func toDeliveryResponse(d *domain.Delivery) api.DeliveryResponse {
	resp := api.DeliveryResponse{
		Id:                    d.ID,
		OrderId:               d.OrderID,
		UserId:                d.UserID,
		Status:                api.DeliveryStatus(d.Status),
		PickupAddress:         d.PickupAddress,
		EstimatedDeliveryDate: d.EstimatedDeliveryDate,
		IsLate:                d.IsLate,
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
	}

	if d.Reschedules != nil {
		reschedules := make([]api.DeliveryRescheduleResponse, 0, len(d.Reschedules))
		for _, r := range d.Reschedules {
			reschedules = append(reschedules, toRescheduleResponse(r))
		}
		resp.Reschedules = &reschedules
	}

	return resp
}

func toRescheduleResponse(r *domain.DeliveryReschedule) api.DeliveryRescheduleResponse {
	return api.DeliveryRescheduleResponse{
		Id:           r.ID,
		PreviousDate: r.PreviousDate,
		NewDate:      r.NewDate,
		Reason:       &r.Reason,
		CreatedAt:    r.CreatedAt,
	}
}
