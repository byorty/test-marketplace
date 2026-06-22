package product

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Price        int64
	Category     string
	Rating       float64
	DeliveryDays int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
