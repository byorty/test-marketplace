package mocks

import (
	"context"
	"sync/atomic"

	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/qr"
	"github.com/google/uuid"
)

type MockDeliveryService struct {
	CreateFn       func(ctx context.Context, input *domain.DeliveryCreateInput) (*domain.Delivery, error)
	GetByIDFn      func(ctx context.Context, userID, id uuid.UUID, role string) (*domain.Delivery, error)
	ListFn         func(ctx context.Context, userID uuid.UUID, role string, filter domain.DeliveryListFilter) (*domain.DeliveryList, error)
	RescheduleFn   func(ctx context.Context, userID, deliveryID uuid.UUID, role string, input *domain.RescheduleInput) (*domain.Delivery, error)
	UpdateStatusFn func(ctx context.Context, userID, deliveryID uuid.UUID, role string, status domain.DeliveryStatus) (*domain.Delivery, error)
	GetQRFn        func(ctx context.Context, userID, deliveryID uuid.UUID, role string) (*qr.QRTokenResponse, error)
	VerifyQRFn     func(ctx context.Context, token string) (*domain.Delivery, error)

	CreateFnCalls       atomic.Int64
	GetByIDFnCalls      atomic.Int64
	ListFnCalls         atomic.Int64
	RescheduleFnCalls   atomic.Int64
	UpdateStatusFnCalls atomic.Int64
	GetQRFnCalls        atomic.Int64
	VerifyQRFnCalls     atomic.Int64
}

func (m *MockDeliveryService) Create(ctx context.Context, input *domain.DeliveryCreateInput) (*domain.Delivery, error) {
	m.CreateFnCalls.Add(1)
	if m.CreateFn == nil {
		panic("CreateFn is nil")
	}
	return m.CreateFn(ctx, input)
}

func (m *MockDeliveryService) GetByID(ctx context.Context, userID, id uuid.UUID, role string) (*domain.Delivery, error) {
	m.GetByIDFnCalls.Add(1)
	if m.GetByIDFn == nil {
		panic("GetByIDFn is nil")
	}
	return m.GetByIDFn(ctx, userID, id, role)
}

func (m *MockDeliveryService) List(ctx context.Context, userID uuid.UUID, role string, filter domain.DeliveryListFilter) (*domain.DeliveryList, error) {
	m.ListFnCalls.Add(1)
	if m.ListFn == nil {
		panic("ListFn is nil")
	}
	return m.ListFn(ctx, userID, role, filter)
}

func (m *MockDeliveryService) Reschedule(ctx context.Context, userID, deliveryID uuid.UUID, role string, input *domain.RescheduleInput) (*domain.Delivery, error) {
	m.RescheduleFnCalls.Add(1)
	if m.RescheduleFn == nil {
		panic("RescheduleFn is nil")
	}
	return m.RescheduleFn(ctx, userID, deliveryID, role, input)
}

func (m *MockDeliveryService) UpdateStatus(ctx context.Context, userID, deliveryID uuid.UUID, role string, status domain.DeliveryStatus) (*domain.Delivery, error) {
	m.UpdateStatusFnCalls.Add(1)
	if m.UpdateStatusFn == nil {
		panic("UpdateStatusFn is nil")
	}
	return m.UpdateStatusFn(ctx, userID, deliveryID, role, status)
}

func (m *MockDeliveryService) GetQR(ctx context.Context, userID, deliveryID uuid.UUID, role string) (*qr.QRTokenResponse, error) {
	m.GetQRFnCalls.Add(1)
	if m.GetQRFn == nil {
		panic("GetQRFn is nil")
	}
	return m.GetQRFn(ctx, userID, deliveryID, role)
}

func (m *MockDeliveryService) VerifyQR(ctx context.Context, token string) (*domain.Delivery, error) {
	m.VerifyQRFnCalls.Add(1)
	if m.VerifyQRFn == nil {
		panic("VerifyQRFn is nil")
	}
	return m.VerifyQRFn(ctx, token)
}
