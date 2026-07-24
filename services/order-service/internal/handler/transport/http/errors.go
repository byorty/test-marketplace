package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain/order"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
	"github.com/byorty/test-marketplace/services/order-service/internal/service"
)

func errorResponse(code, message string) api.ErrorResponse {
	return api.ErrorResponse{
		Code: code,
		Message: message,
	}
}

func mapAddToCartError(log *slog.Logger, err error) api.AddToCartResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return api.AddToCart400JSONResponse(
			errorResponse("validation_error", err.Error()),
		)

	default:
		log.Error("add product to cart failed", 
		slog.String("op", "Handler.AddToCart"),
		slog.Any("error", err))

		return api.AddToCart500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapGetCartError(log *slog.Logger, err error) api.GetCartResponseObject {
	log.Error("get cart failed", 
	slog.String("op", "Handler.GetCart"),
	slog.Any("error", err))

	return api.GetCart500JSONResponse(
		errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
	)
}

func mapRemoveFromCartError(log *slog.Logger, err error) api.RemoveFromCartResponseObject {
	switch {
	case errors.Is(err, order.ErrCartItemNotFound):
		return api.RemoveFromCart404JSONResponse(
			errorResponse("product_not_found", err.Error()),
		)
	default:
		log.Error("remove from cart failed", 
		slog.String("op", "Handler.RemoveFromCart"),
		slog.Any("error", err))
		
		return api.RemoveFromCart500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapGetOrderByIDError(log *slog.Logger, err error) api.GetOrderByIDResponseObject {
	switch {
	case errors.Is(err, order.ErrOrderNotFound):
		return api.GetOrderByID404JSONResponse(
			errorResponse("order_not_found", err.Error()),
		)

	default:
		log.Error("get order by id failed",
		slog.String("op", "Handler.GetOrderByID"),
		slog.Any("error", err))

		return api.GetOrderByID500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapCreateOrderError(log *slog.Logger, err error) api.CreateOrderResponseObject {
	switch {
	case errors.Is(err, order.ErrCartEmpty):
		return api.CreateOrder400JSONResponse(
			errorResponse("validation_error", err.Error()),
		)

	default:
		log.Error("create order failed",
		slog.String("op", "Handler.CreateOrder"),
		slog.Any("error", err))

		return api.CreateOrder500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}