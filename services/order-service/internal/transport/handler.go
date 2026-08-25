package transport

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"go.uber.org/zap"
)


type OrderHandler struct {
    service    domain.OrderService
    log        *zap.Logger
    authorizer domain.Authorizer
}

func New(service domain.OrderService, log *zap.Logger, authorizer domain.Authorizer) *OrderHandler {
    return &OrderHandler{
        service:    service,
        log:        log.Named("order-handler"),
        authorizer: authorizer,
    }
}