package service

import "errors"

var (
	ErrInvalidID            = errors.New("invalid id")
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidOrderID       = errors.New("invalid order id")
	ErrInvalidUserID        = errors.New("invalid user id")
	ErrInvalidPickupAddress = errors.New("invalid pickup address")
	ErrInvalidEstimatedDate = errors.New("invalid estimated delivery date")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrInvalidNewDate       = errors.New("invalid new date")
	ErrNilInput             = errors.New("input is nil")
	ErrForbidden            = errors.New("forbidden")
	ErrOrderNotAccessible   = errors.New("order is not accessible")
)
