package domain

import (
	"context"

	"github.com/google/uuid"
)

type DeliveryRepository interface {
	Create(ctx context.Context, delivery *Delivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*Delivery, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*Delivery, error)
	List(ctx context.Context, filter DeliveryListFilter) (*DeliveryList, error)
	Update(ctx context.Context, delivery *Delivery) (*Delivery, error)
	UpdateStatusConditional(ctx context.Context, id uuid.UUID, fromStatus DeliveryStatus, toStatus DeliveryStatus, updates map[string]interface{}) (*Delivery, error)
	CreateReschedule(ctx context.Context, reschedule *DeliveryReschedule) error
	ListReschedules(ctx context.Context, deliveryID uuid.UUID) ([]DeliveryReschedule, error)
	RescheduleTx(ctx context.Context, delivery *Delivery, reschedule *DeliveryReschedule) (*Delivery, error)
}
