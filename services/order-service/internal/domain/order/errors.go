package order

import "errors"

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrCartEmpty = errors.New("cart is empty")
	ErrProductAlreadyInCart = errors.New("product already in cart")
	ErrProductNotInCart = errors.New("product not found in cart")
)