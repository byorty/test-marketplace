package service

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNilInput = errors.New("input is nil")
	ErrInvalidID = errors.New("invalid id")
	ErrBadFilter = errors.New("invalid filter")
	ErrInvalidProductName = errors.New("invalid product name")
	ErrEmptyUpdate = errors.New("update fields are empty")
)