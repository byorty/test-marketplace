package transport

import (
	"errors"
	"net/http"

	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/qr"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/service"
	"go.uber.org/zap"
)

func errorResponse(code, message string) api.Error {
	return api.Error{Code: code, Message: message}
}

func mapCreateDeliveryError(log *zap.Logger, err error) api.CreateDeliveryResponseObject {
	switch {
	case errors.Is(err, service.ErrNilInput),
		errors.Is(err, service.ErrInvalidOrderID),
		errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidInput),
		errors.Is(err, service.ErrInvalidEstimatedDate):
		return api.CreateDelivery400JSONResponse(errorResponse("validation_error", err.Error()))
	case errors.Is(err, domain.ErrAccessDenied):
		return api.CreateDelivery403JSONResponse(errorResponse("forbidden", err.Error()))
	case errors.Is(err, domain.ErrOrderNotPaid):
		return api.CreateDelivery400JSONResponse(errorResponse("order_not_paid", err.Error()))
	case errors.Is(err, domain.ErrDeliveryAlreadyExists):
		return api.CreateDelivery409JSONResponse(errorResponse("delivery_already_exists", err.Error()))
	default:
		log.Error("create delivery failed", zap.Error(err))
		return api.CreateDelivery500JSONResponse(errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)))
	}
}

func mapGetDeliveriesError(log *zap.Logger, err error) api.GetDeliveriesResponseObject {
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.GetDeliveries403JSONResponse(errorResponse("forbidden", err.Error()))
	default:
		log.Error("list deliveries failed", zap.Error(err))
		return api.GetDeliveries500JSONResponse(errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)))
	}
}

func mapGetDeliveryError(log *zap.Logger, err error) api.GetDeliveryResponseObject {
	switch {
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return api.GetDelivery404JSONResponse(errorResponse("delivery_not_found", err.Error()))
	case errors.Is(err, domain.ErrAccessDenied):
		return api.GetDelivery403JSONResponse(errorResponse("forbidden", err.Error()))
	default:
		log.Error("get delivery failed", zap.Error(err))
		return api.GetDelivery500JSONResponse(errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)))
	}
}

func mapRescheduleDeliveryError(log *zap.Logger, err error) api.RescheduleDeliveryResponseObject {
	switch {
	case errors.Is(err, service.ErrNilInput),
		errors.Is(err, service.ErrInvalidNewDate):
		return api.RescheduleDelivery400JSONResponse(errorResponse("validation_error", err.Error()))
	case errors.Is(err, domain.ErrAccessDenied):
		return api.RescheduleDelivery403JSONResponse(errorResponse("forbidden", err.Error()))
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return api.RescheduleDelivery404JSONResponse(errorResponse("delivery_not_found", err.Error()))
	case errors.Is(err, domain.ErrInvalidDeliveryStatus):
		return api.RescheduleDelivery400JSONResponse(errorResponse("invalid_delivery_status", err.Error()))
	case errors.Is(err, domain.ErrRescheduleOnTerminal):
		return api.RescheduleDelivery409JSONResponse(errorResponse("cannot_reschedule", err.Error()))
	default:
		log.Error("reschedule delivery failed", zap.Error(err))
		return api.RescheduleDelivery500JSONResponse(errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)))
	}
}

func mapUpdateDeliveryStatusError(log *zap.Logger, err error) api.UpdateDeliveryStatusResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidStatus):
		return api.UpdateDeliveryStatus400JSONResponse(errorResponse("validation_error", err.Error()))
	case errors.Is(err, domain.ErrAccessDenied):
		return api.UpdateDeliveryStatus403JSONResponse(errorResponse("forbidden", err.Error()))
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return api.UpdateDeliveryStatus404JSONResponse(errorResponse("delivery_not_found", err.Error()))
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		return api.UpdateDeliveryStatus400JSONResponse(errorResponse("invalid_status_transition", err.Error()))
	case errors.Is(err, domain.ErrDeliveryAlreadyDelivered):
		return api.UpdateDeliveryStatus409JSONResponse(errorResponse("delivery_already_delivered", err.Error()))
	case errors.Is(err, domain.ErrDeliveryAlreadyCancelled):
		return api.UpdateDeliveryStatus409JSONResponse(errorResponse("delivery_already_cancelled", err.Error()))
	default:
		log.Error("update delivery status failed", zap.Error(err))
		return api.UpdateDeliveryStatus500JSONResponse(errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)))
	}
}

func mapGetDeliveryQRError(log *zap.Logger, err error) api.GetDeliveryQRResponseObject {
	switch {
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return api.GetDeliveryQR404JSONResponse(errorResponse("delivery_not_found", err.Error()))
	case errors.Is(err, domain.ErrAccessDenied):
		return api.GetDeliveryQR403JSONResponse(errorResponse("forbidden", err.Error()))
	case errors.Is(err, domain.ErrInvalidDeliveryStatus), errors.Is(err, domain.ErrQRNotAvailable):
		return api.GetDeliveryQR409JSONResponse(errorResponse("invalid_delivery_status", err.Error()))
	default:
		log.Error("get delivery qr failed", zap.Error(err))
		return api.GetDeliveryQR500JSONResponse(errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)))
	}
}

func mapVerifyDeliveryQRError(log *zap.Logger, err error) api.VerifyDeliveryQRResponseObject {
	switch {
	case errors.Is(err, qr.ErrInvalidSignature),
		errors.Is(err, qr.ErrTokenExpired):
		return api.VerifyDeliveryQR401JSONResponse(errorResponse("unauthorized", err.Error()))
	case errors.Is(err, qr.ErrInvalidToken),
		errors.Is(err, qr.ErrTokenRevoked),
		errors.Is(err, domain.ErrInvalidDeliveryStatus),
		errors.Is(err, domain.ErrQRNotAvailable),
		errors.Is(err, domain.ErrDeliveryAlreadyDelivered):
		return api.VerifyDeliveryQR400JSONResponse(errorResponse("invalid_token", err.Error()))
	case errors.Is(err, domain.ErrAccessDenied):
		return api.VerifyDeliveryQR403JSONResponse(errorResponse("forbidden", err.Error()))
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return api.VerifyDeliveryQR404JSONResponse(errorResponse("delivery_not_found", err.Error()))
	default:
		log.Error("verify delivery qr failed", zap.Error(err))
		return api.VerifyDeliveryQR500JSONResponse(errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)))
	}
}
