package product

import "github.com/google/uuid"

type Product struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Price         int64     `json:"price"`
	DeliveryDays  int32     `json:"delivery_days"`
	Rating        float64   `json:"rating"`
}