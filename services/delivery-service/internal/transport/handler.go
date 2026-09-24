package transport

import (
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
	"go.uber.org/zap"
)

type DeliveryHandler struct {
	service    domain.DeliveryService
	log        *zap.Logger
	authorizer domain.Authorizer
}

func NewDeliveryHandler(service domain.DeliveryService, log *zap.Logger, authorizer domain.Authorizer) *DeliveryHandler {
	return &DeliveryHandler{
		service:    service,
		log:        log.Named("delivery-handler"),
		authorizer: authorizer,
	}
}

var _ api.StrictServerInterface = (*DeliveryHandler)(nil)
