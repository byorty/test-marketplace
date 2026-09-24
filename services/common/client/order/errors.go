package order

import "errors"

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrOrderNotPaid   = errors.New("order is not paid")
	ErrForbidden      = errors.New("access denied to order")
	ErrNotImplemented = errors.New("not implemented")
)
