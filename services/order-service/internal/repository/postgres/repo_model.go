package postgres

import (
	"time"

	"github.com/google/uuid"
)

type OrderModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID
	Status string
	Total int64
	CreatedAt time.Time
	DeliveryDate time.Time
	Items []OrderItemModel
}

type OrderItemModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrderID uuid.UUID
	ProductID uuid.UUID
	ProductName string
	ProductPrice int64
	Quantity int
}

type CartItemModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID
	ProductID uuid.UUID
	Quantity int
}