package client

import (
	"context"

	client "github.com/byorty/test-marketplace/services/common/client/product/generated"
	"github.com/google/uuid"
)

type Client interface {
	GetProduct(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error)
}