package qr

import (
	"time"

	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type QRToken struct {
	DeliveryID uuid.UUID `json:"delivery_id"`
	OrderID    uuid.UUID `json:"order_id"`
	UserID     uuid.UUID `json:"user_id"`
	Nonce      uuid.UUID `json:"nonce"`
	IssuedAt   int64     `json:"iat"`
	ExpiresAt  int64     `json:"exp"`
}

type QRTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
