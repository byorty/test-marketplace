package order

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID uuid.UUID
	UserID uuid.UUID
	Status string
	Total int64
	CreatedAt time.Time
	DeliveryDate time.Time
	Items []OrderItem
}

type OrderItem struct {
	ID uuid.UUID
	OrderID uuid.UUID
	ProductID uuid.UUID
	ProductName string
	ProductPrice int64
	Quantity int
}

type CartItem struct {
	ID uuid.UUID
	UserID uuid.UUID
	ProductID uuid.UUID
	Quantity int
}