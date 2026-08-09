package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Product struct {
	bun.BaseModel `bun:"table:products"`

	ID           uuid.UUID `bun:"id,pk,type:uuid"`
	Name         string `bun:"name,notnull"`
	Description  string `bun:"description"`
	Price        int64 `bun:"price,notnull"`
	Category     string `bun:"category,notnull"`
	Rating       float64 `bun:"rating,notnull"`
	DeliveryDays int `bun:"delivery_days,notnull"`
	CreatedAt    time.Time `bun:"created_at,notnull"`
	UpdatedAt    time.Time `bun:"updated_at,notnull"`
}
