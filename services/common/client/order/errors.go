package order

import "errors"

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrOrderNotPaid   = errors.New("order is not paid")
	ErrNotImplemented = errors.New("not implemented")
)
