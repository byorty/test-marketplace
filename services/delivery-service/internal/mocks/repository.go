package mocks

import (
	"context"
	"sync/atomic"

	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"github.com/google/uuid"
)

type MockDeliveryRepository struct {
	CreateFn                  func(ctx context.Context, delivery *domain.Delivery) error
	GetByIDFn                 func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error)
	GetByOrderIDFn            func(ctx context.Context, orderID uuid.UUID) (*domain.Delivery, error)
	ListFn                    func(ctx context.Context, filter domain.DeliveryListFilter) (*domain.DeliveryList, error)
	UpdateFn                  func(ctx context.Context, delivery *domain.Delivery) (*domain.Delivery, error)
	UpdateStatusConditionalFn func(ctx context.Context, id uuid.UUID, fromStatus domain.DeliveryStatus, toStatus domain.DeliveryStatus, updates map[string]interface{}) (*domain.Delivery, error)
	CreateRescheduleFn        func(ctx context.Context, reschedule *domain.DeliveryReschedule) error
	ListReschedulesFn         func(ctx context.Context, deliveryID uuid.UUID) ([]domain.DeliveryReschedule, error)
	RescheduleTxFn            func(ctx context.Context, delivery *domain.Delivery, reschedule *domain.DeliveryReschedule) (*domain.Delivery, error)

	CreateFnCalls                  atomic.Int64
	GetByIDFnCalls                 atomic.Int64
	GetByOrderIDFnCalls            atomic.Int64
	ListFnCalls                    atomic.Int64
	UpdateFnCalls                  atomic.Int64
	UpdateStatusConditionalFnCalls atomic.Int64
	CreateRescheduleFnCalls        atomic.Int64
	ListReschedulesFnCalls         atomic.Int64
	RescheduleTxFnCalls            atomic.Int64
}

func (m *MockDeliveryRepository) Create(ctx context.Context, delivery *domain.Delivery) error {
	m.CreateFnCalls.Add(1)
	if m.CreateFn == nil {
		panic("CreateFn is nil")
	}
	return m.CreateFn(ctx, delivery)
}

func (m *MockDeliveryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	m.GetByIDFnCalls.Add(1)
	if m.GetByIDFn == nil {
		panic("GetByIDFn is nil")
	}
	return m.GetByIDFn(ctx, id)
}

func (m *MockDeliveryRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Delivery, error) {
	m.GetByOrderIDFnCalls.Add(1)
	if m.GetByOrderIDFn == nil {
		panic("GetByOrderIDFn is nil")
	}
	return m.GetByOrderIDFn(ctx, orderID)
}

func (m *MockDeliveryRepository) List(ctx context.Context, filter domain.DeliveryListFilter) (*domain.DeliveryList, error) {
	m.ListFnCalls.Add(1)
	if m.ListFn == nil {
		panic("ListFn is nil")
	}
	return m.ListFn(ctx, filter)
}

func (m *MockDeliveryRepository) Update(ctx context.Context, delivery *domain.Delivery) (*domain.Delivery, error) {
	m.UpdateFnCalls.Add(1)
	if m.UpdateFn == nil {
		panic("UpdateFn is nil")
	}
	return m.UpdateFn(ctx, delivery)
}

func (m *MockDeliveryRepository) UpdateStatusConditional(ctx context.Context, id uuid.UUID, fromStatus domain.DeliveryStatus, toStatus domain.DeliveryStatus, updates map[string]interface{}) (*domain.Delivery, error) {
	m.UpdateStatusConditionalFnCalls.Add(1)
	if m.UpdateStatusConditionalFn == nil {
		panic("UpdateStatusConditionalFn is nil")
	}
	return m.UpdateStatusConditionalFn(ctx, id, fromStatus, toStatus, updates)
}

func (m *MockDeliveryRepository) CreateReschedule(ctx context.Context, reschedule *domain.DeliveryReschedule) error {
	m.CreateRescheduleFnCalls.Add(1)
	if m.CreateRescheduleFn == nil {
		panic("CreateRescheduleFn is nil")
	}
	return m.CreateRescheduleFn(ctx, reschedule)
}

func (m *MockDeliveryRepository) ListReschedules(ctx context.Context, deliveryID uuid.UUID) ([]domain.DeliveryReschedule, error) {
	m.ListReschedulesFnCalls.Add(1)
	if m.ListReschedulesFn == nil {
		panic("ListReschedulesFn is nil")
	}
	return m.ListReschedulesFn(ctx, deliveryID)
}

func (m *MockDeliveryRepository) RescheduleTx(ctx context.Context, delivery *domain.Delivery, reschedule *domain.DeliveryReschedule) (*domain.Delivery, error) {
	m.RescheduleTxFnCalls.Add(1)
	if m.RescheduleTxFn == nil {
		panic("RescheduleTxFn is nil")
	}
	return m.RescheduleTxFn(ctx, delivery, reschedule)
}
