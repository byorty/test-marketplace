package domain

import "context"

type Notifier interface {
	NotifyDeliveryLate(ctx context.Context, delivery *Delivery) error
	NotifyDeliveryStatusChanged(ctx context.Context, delivery *Delivery, oldStatus DeliveryStatus) error
}
