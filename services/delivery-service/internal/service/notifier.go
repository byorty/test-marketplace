package service

import (
	"context"

	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"go.uber.org/zap"
)

type LoggingNotifier struct {
	log *zap.Logger
}

func NewLoggingNotifier(log *zap.Logger) domain.Notifier {
	return &LoggingNotifier{log: log.Named("notifier")}
}

func (n *LoggingNotifier) NotifyDeliveryLate(ctx context.Context, delivery *domain.Delivery) error {
	n.log.Info("delivery is late",
		zap.String("delivery_id", delivery.ID.String()),
		zap.String("order_id", delivery.OrderID.String()),
		zap.String("user_id", delivery.UserID.String()),
		zap.Time("estimated_delivery_date", delivery.EstimatedDeliveryDate),
	)
	return nil
}

func (n *LoggingNotifier) NotifyDeliveryStatusChanged(ctx context.Context, delivery *domain.Delivery, oldStatus domain.DeliveryStatus) error {
	n.log.Info("delivery status changed",
		zap.String("delivery_id", delivery.ID.String()),
		zap.String("old_status", string(oldStatus)),
		zap.String("new_status", string(delivery.Status)),
	)
	return nil
}
