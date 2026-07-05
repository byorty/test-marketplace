package http

import (
	"errors"
	"log/slog"
	"net/http"

	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	"github.com/byorty/test-marketplace/services/product-service/internal/service"
)

func errorResponse(code, message string) api.Error {
	return api.Error{
		Code: code,
		Message: message,
	}
}

func mapCreateError(log *slog.Logger, err error) api.CreateProductResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return api.CreateProduct400JSONResponse(
			errorResponse("validation_errror", err.Error()),
		)
		
	default:
		log.Error("create product failed", slog.Any("error", err))

		return api.CreateProduct500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapGetError(log *slog.Logger, err error) api.GetByIDResponseObject {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return api.GetByID404JSONResponse(
			errorResponse("product_not_found", err.Error()),
		)

	default:
		log.Error("get product failed", slog.Any("error", err))

		return api.GetByID500JSONResponse(
			errorResponse("internal error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapUpdateError(log *slog.Logger, err error) api.UpdateProductResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return api.UpdateProduct400JSONResponse(
			errorResponse("validation_error", err.Error()),
		)

	case errors.Is(err, service.ErrEmptyUpdate):
		return api.UpdateProduct400JSONResponse(
			errorResponse("empty_update", err.Error()),
		)

	case errors.Is(err, domain.ErrProductNotFound):
		return api.UpdateProduct404JSONResponse(
			errorResponse("product_not_found", err.Error()),
		)

	default:
		log.Error("update product failed", slog.Any("error", err))
	
		return api.UpdateProduct500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapDeleteError(log *slog.Logger, err error) api.DeleteProductResponseObject {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return api.DeleteProduct404JSONResponse(
			errorResponse("product_not_found", err.Error()),
		)

	default:
		log.Error("delete product failed", slog.Any("error", err))
	
		return api.DeleteProduct500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapListError(log *slog.Logger, err error) api.GetProductsResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return api.GetProducts400JSONResponse(
			errorResponse("validation_error", err.Error()),
		)

	default:
		log.Error("list products failed", slog.Any("error", err))

		return api.GetProducts500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}