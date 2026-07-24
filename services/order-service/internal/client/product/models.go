package product

import "github.com/google/uuid"

type Product struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
	Price int64 `json:"price"`
	DeliveryDays int `json:"delivery_days"`
}