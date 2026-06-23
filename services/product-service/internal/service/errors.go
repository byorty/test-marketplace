package service

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrBadFilter = errors.New("invalid filter")
	ErrInvalidProductName = errors.New("invalid product name")
	ErrEmptyUpdate = errors.New("update fields are empty")
)