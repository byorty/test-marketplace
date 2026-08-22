package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Status string

const (
	StatusCreated Status = "CREATED"
	StatusPaid Status = "PAID"
	StatusDelivering Status = "DELIVERING"
	StatusDelivered Status = "DELIVERED"
)
type Order struct {
	bun.BaseModel `bun:"table:orders"`

	ID           uuid.UUID `bun:"id,pk,type:uuid"`
	UserID       uuid.UUID `bun:"user_id,notnull,type:uuid"`
	Status       Status    `bun:"status,notnull,type:varchar(20)"`
	TotalPrice        int64     `bun:"total_price,notnull"`
	CreatedAt    time.Time `bun:"created_at,notnull"`
	DeliveryDate time.Time `bun:"delivery_date"`
	Items        []OrderItem `bun:"rel:has-many,join:id=order_id"`
}

type OrderItem struct {
	bun.BaseModel `bun:"table:order_items"`

	ID           uuid.UUID `bun:"id,pk,type:uuid"`
	OrderID      uuid.UUID `bun:"order_id,notnull,type:uuid"`
	ProductID    uuid.UUID `bun:"product_id,notnull,type:uuid"`
	ProductPrice int64     `bun:"product_price,notnull"`
	Quantity     int       `bun:"quantity,notnull"`
}

type CartItem struct {
	bun.BaseModel `bun:"table:cart_items"`

	ID        uuid.UUID `bun:"id,pk,type:uuid"`
	UserID    uuid.UUID `bun:"user_id,notnull,type:uuid"`
	ProductID uuid.UUID `bun:"product_id,notnull,type:uuid"`
	Quantity  int       `bun:"quantity,notnull"`
}

type Cart struct {
	Items      []CartItem
	TotalPrice int64
}