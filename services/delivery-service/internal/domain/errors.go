package domain

import "errors"

var (
	ErrDeliveryNotFound         = errors.New("delivery not found")
	ErrDeliveryAlreadyExists    = errors.New("delivery already exists for this order")
	ErrInvalidDeliveryStatus    = errors.New("invalid delivery status")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrDeliveryAlreadyDelivered = errors.New("delivery already delivered")
	ErrDeliveryAlreadyCancelled = errors.New("delivery already cancelled")
	ErrRescheduleOnTerminal     = errors.New("cannot reschedule delivery in terminal status")
	ErrQRNotAvailable           = errors.New("QR code is not available for this delivery status")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderNotPaid             = errors.New("order is not in PAID status")
	ErrAccessDenied             = errors.New("access denied")
)
