package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type DeliveryReschedule struct {
	bun.BaseModel `bun:"table:delivery_reschedules"`

	ID           uuid.UUID `bun:"id,pk,type:uuid"`
	DeliveryID   uuid.UUID `bun:"delivery_id,notnull,type:uuid"`
	PreviousDate time.Time `bun:"previous_date,notnull"`
	NewDate      time.Time `bun:"new_date,notnull"`
	Reason       string    `bun:"reason"`
	CreatedAt    time.Time `bun:"created_at,notnull"`
}

type RescheduleInput struct {
	NewDate time.Time
	Reason  string
}
