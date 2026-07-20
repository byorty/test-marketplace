package postgres

import (
	"time"

	"github.com/google/uuid"
)

type ProductModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string
	Description string
	Category string
	Price int64
	DeliveryDays int
	Rating float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
