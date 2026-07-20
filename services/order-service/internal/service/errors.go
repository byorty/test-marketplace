package service

import "errors"

var (
	ErrInvalidID = errors.New("invalid id")
	ErrInvalidUserID = errors.New("invalid user id")
	ErrInvaliProductdID = errors.New("invalid product id")
	ErrInvalidOrderID = errors.New("invalid order id")
	ErrInvalidQuantity = errors.New("invalid quantity")
	ErrInvalidInput = errors.New("invalid input")
	ErrInvalidPrice = errors.New("invalid price")
	ErrEmptyOrder = errors.New("order is empty")
)