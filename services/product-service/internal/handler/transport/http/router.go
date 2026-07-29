package http

import (
	"net/http"

	"github.com/byorty/test-marketplace/services/product-service/internal/auth"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	"github.com/byorty/test-marketplace/services/product-service/internal/handler/transport/http/middlwr"
	"github.com/byorty/test-marketplace/services/product-service/internal/rbac"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *Handler, jwt *auth.Validator, authorizer *rbac.Authorizer) http.Handler {
	router := chi.NewRouter()

	router.Use(middlwr.NewAuth(jwt).Handler)
	router.Use(middlwr.NewAuthorization(authorizer).Handler)

	strictHandler := api.NewStrictHandler(handler, nil)

	api.HandlerFromMux(strictHandler, router)

	return router
}