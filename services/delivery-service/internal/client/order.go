package client

import (
	"context"

	orderclient "github.com/byorty/test-marketplace/services/common/client/order"
	"github.com/google/uuid"
)

type Client interface {
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*orderclient.OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error
}
