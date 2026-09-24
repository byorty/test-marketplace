package service

import (
	"time"

	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
)

type RealClock struct{}

func NewRealClock() domain.Clock {
	return &RealClock{}
}

func (c *RealClock) Now() time.Time {
	return time.Now()
}
