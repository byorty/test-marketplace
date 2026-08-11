package app

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

func NewValidator() *validator.Validate {
	return validator.New()
}

var ValidateModule = fx.Provide(NewValidator)