package domain

import (
	"context"

	"github.com/byorty/test-marketplace/services/delivery-service/internal/qr"
	"github.com/google/uuid"
)

type DeliveryService interface {
	Create(ctx context.Context, input *DeliveryCreateInput) (*Delivery, error)
	GetByID(ctx context.Context, userID, id uuid.UUID, role string) (*Delivery, error)
	List(ctx context.Context, userID uuid.UUID, role string, filter DeliveryListFilter) (*DeliveryList, error)
	Reschedule(ctx context.Context, userID, deliveryID uuid.UUID, role string, input *RescheduleInput) (*Delivery, error)
	UpdateStatus(ctx context.Context, userID, deliveryID uuid.UUID, role string, status DeliveryStatus) (*Delivery, error)
	GetQR(ctx context.Context, userID, deliveryID uuid.UUID, role string) (*qr.QRTokenResponse, error)
	VerifyQR(ctx context.Context, token string) (*Delivery, error)
}
