package order

import "errors"

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrCartEmpty = errors.New("cart is empty")
	ErrProductNotInCart = errors.New("product not found in cart")
	ErrCartItemNotFound = errors.New("item not found")
)