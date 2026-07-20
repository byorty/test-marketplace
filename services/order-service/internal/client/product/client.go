package product

import (
	"context"

	"github.com/google/uuid"
)

type Product struct {
	ID uuid.UUID
	Name string
	Price int64
	DeliveryDays int
}

type Client interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
}