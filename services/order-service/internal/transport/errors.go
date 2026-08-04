package transport

import (
	"errors"
	"net/http"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
	"github.com/byorty/test-marketplace/services/order-service/internal/service"
	"go.uber.org/zap"
)

func errorResponse(code, message string) api.ErrorResponse {
	return api.ErrorResponse{
		Code: code,
		Message: message,
	}
}

func mapAddToCartError(log *zap.Logger, err error) api.AddToCartResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidInput),
		errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidProductID),
		errors.Is(err, service.ErrInvalidQuantity):

		return api.AddToCart400JSONResponse(
			errorResponse("validation_error", err.Error()),
		)

	case errors.Is(err, service.ErrProductNotFound):
		return api.AddToCart404JSONResponse(
			errorResponse("product_not_found", err.Error()),
		)

	default:
		log.Error(
			"add product to cart failed",
			zap.Error(err),
		)

		return api.AddToCart500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapGetCartError(log *zap.Logger, err error) api.GetCartResponseObject {
	log.Error(
			"get cart failed",
			zap.Error(err),
		)

	return api.GetCart500JSONResponse(
		errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
	)
}

func mapRemoveFromCartError(log *zap.Logger, err error) api.RemoveFromCartResponseObject {
	switch {

	case errors.Is(err, domain.ErrCartItemNotFound):
		return api.RemoveFromCart404JSONResponse(
			errorResponse("cart_item_not_found", err.Error()),
		)

	default:
		log.Error(
			"remove from cart failed",
			zap.Error(err),
		)

		return api.RemoveFromCart500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapGetOrderByIDError(log *zap.Logger, err error) api.GetOrderByIDResponseObject {
	switch {

	case errors.Is(err, domain.ErrOrderNotFound):
		return api.GetOrderByID404JSONResponse(
			errorResponse("order_not_found", err.Error()),
		)

	default:
		log.Error(
			"get order by id failed",
			zap.Error(err),
		)

		return api.GetOrderByID500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapCreateOrderError(log *zap.Logger, err error) api.CreateOrderResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidUserID):
		return api.CreateOrder400JSONResponse(
			errorResponse("validation_error", err.Error()),
		)

	case errors.Is(err, domain.ErrCartEmpty):
		return api.CreateOrder400JSONResponse(
			errorResponse("cart_empty", err.Error()),
		)

	default:
		log.Error(
			"create order failed",
			zap.Error(err),
		)

		return api.CreateOrder500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}