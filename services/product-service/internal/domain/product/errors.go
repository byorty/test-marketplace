package product

import "errors"

var (
	ErrNotFound = errors.New("product not found")
	ErrInvalidName = errors.New("invalid product name")
	ErrInvalidPrice = errors.New("invalid product price")
	ErrInvalidRating = errors.New("invalid product rating")
	ErrInvalidDeliveryDays = errors.New("invalid delivery days")
)