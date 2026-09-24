package mocks

import (
	"context"
	"sync/atomic"

	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
)

type MockNotifier struct {
	NotifyDeliveryLateFn             func(ctx context.Context, delivery *domain.Delivery) error
	NotifyDeliveryStatusChangedFn    func(ctx context.Context, delivery *domain.Delivery, oldStatus domain.DeliveryStatus) error
	NotifyDeliveryLateCalls          atomic.Int64
	NotifyDeliveryStatusChangedCalls atomic.Int64
}

func (m *MockNotifier) NotifyDeliveryLate(ctx context.Context, delivery *domain.Delivery) error {
	m.NotifyDeliveryLateCalls.Add(1)
	if m.NotifyDeliveryLateFn == nil {
		return nil
	}
	return m.NotifyDeliveryLateFn(ctx, delivery)
}

func (m *MockNotifier) NotifyDeliveryStatusChanged(ctx context.Context, delivery *domain.Delivery, oldStatus domain.DeliveryStatus) error {
	m.NotifyDeliveryStatusChangedCalls.Add(1)
	if m.NotifyDeliveryStatusChangedFn == nil {
		return nil
	}
	return m.NotifyDeliveryStatusChangedFn(ctx, delivery, oldStatus)
}
