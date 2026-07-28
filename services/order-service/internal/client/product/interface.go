package product

import (
	"context"

	"github.com/google/uuid"
)

type Client interface {
	GetProduct(ctx context.Context, id uuid.UUID) (*Product, error)
}