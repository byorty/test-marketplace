package transport

import (
	"errors"
	"net/http"

	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	"github.com/byorty/test-marketplace/services/product-service/internal/service"
	"go.uber.org/zap"
)

func errorResponse(code, message string) api.Error {
	return api.Error{
		Code: code,
		Message: message,
	}
}

func mapCreateError(log *zap.Logger, err error) api.CreateProductResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return api.CreateProduct400JSONResponse(
			errorResponse("validation_error", err.Error()),
		)

	default:
		log.Error(
			"create product failed",
			zap.Error(err),
		)

		return api.CreateProduct500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapGetError(log *zap.Logger, err error) api.GetByIDResponseObject {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return api.GetByID404JSONResponse(
			errorResponse("product_not_found", err.Error()),
		)

	default:
		log.Error(
			"get product failed",
			zap.Error(err),
		)

		return api.GetByID500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapUpdateError(log *zap.Logger, err error) api.UpdateProductResponseObject {
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
		log.Error(
			"update product failed",
			zap.Error(err),
		)

		return api.UpdateProduct500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapDeleteError(log *zap.Logger, err error) api.DeleteProductResponseObject {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return api.DeleteProduct404JSONResponse(
			errorResponse("product_not_found", err.Error()),
		)

	default:
		log.Error(
			"delete product failed",
			zap.Error(err),
		)

		return api.DeleteProduct500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}

func mapListError(log *zap.Logger, err error) api.GetProductsResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return api.GetProducts400JSONResponse(
			errorResponse("validation_error", err.Error()),
		)

	default:
		log.Error(
			"list products failed",
			zap.Error(err),
		)

		return api.GetProducts500JSONResponse(
			errorResponse("internal_error", http.StatusText(http.StatusInternalServerError)),
		)
	}
}