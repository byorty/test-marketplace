package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type DeliveryStatus string

const (
	DeliveryStatusPending        DeliveryStatus = "PENDING"
	DeliveryStatusInTransit      DeliveryStatus = "IN_TRANSIT"
	DeliveryStatusReadyForPickup DeliveryStatus = "READY_FOR_PICKUP"
	DeliveryStatusDelivered      DeliveryStatus = "DELIVERED"
	DeliveryStatusCancelled      DeliveryStatus = "CANCELLED"
)

func (s DeliveryStatus) IsValid() bool {
	switch s {
	case DeliveryStatusPending,
		DeliveryStatusInTransit,
		DeliveryStatusReadyForPickup,
		DeliveryStatusDelivered,
		DeliveryStatusCancelled:
		return true
	}
	return false
}

func (s DeliveryStatus) IsTerminal() bool {
	return s == DeliveryStatusDelivered || s == DeliveryStatusCancelled
}

func CanTransition(from, to DeliveryStatus) bool {
	switch from {
	case DeliveryStatusPending:
		return to == DeliveryStatusInTransit || to == DeliveryStatusCancelled
	case DeliveryStatusInTransit:
		return to == DeliveryStatusReadyForPickup || to == DeliveryStatusCancelled
	case DeliveryStatusReadyForPickup:
		return to == DeliveryStatusDelivered || to == DeliveryStatusCancelled
	case DeliveryStatusDelivered, DeliveryStatusCancelled:
		return false
	}
	return false
}

func CanReschedule(status DeliveryStatus) bool {
	return status == DeliveryStatusPending || status == DeliveryStatusInTransit
}

func IsOverdue(d *Delivery, now time.Time) bool {
	if d.Status.IsTerminal() {
		return false
	}
	return now.After(d.EstimatedDeliveryDate)
}

type Delivery struct {
	bun.BaseModel `bun:"table:deliveries"`

	ID                    uuid.UUID             `bun:"id,pk,type:uuid"`
	OrderID               uuid.UUID             `bun:"order_id,notnull,type:uuid"`
	UserID                uuid.UUID             `bun:"user_id,notnull,type:uuid"`
	Status                DeliveryStatus        `bun:"status,notnull,type:varchar(30)"`
	PickupAddress         string                `bun:"pickup_address,notnull"`
	EstimatedDeliveryDate time.Time             `bun:"estimated_delivery_date,notnull"`
	IsLate                bool                  `bun:"is_late,notnull,default:false"`
	QRNonce               uuid.NullUUID         `bun:"qr_nonce,type:uuid"`
	Reschedules           []*DeliveryReschedule `bun:"-" json:"reschedules,omitempty"`
	CreatedAt             time.Time             `bun:"created_at,notnull"`
	UpdatedAt             time.Time             `bun:"updated_at,notnull"`
}

type DeliveryList struct {
	Items    []*Delivery
	Total    int64
	Page     int
	PageSize int
}

type DeliveryListFilter struct {
	Status   *DeliveryStatus
	IsLate   *bool
	OrderID  *uuid.UUID
	UserID   *uuid.UUID
	Page     int
	PageSize int
}

type DeliveryCreateInput struct {
	UserID                uuid.UUID
	OrderID               uuid.UUID
	PickupAddress         string
	EstimatedDeliveryDate time.Time
	Role                  string
}
