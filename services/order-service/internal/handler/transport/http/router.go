package http

import (
	"net/http"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
	"github.com/byorty/test-marketplace/services/order-service/internal/handler/transport/http/middlwr"
	"github.com/byorty/test-marketplace/services/order-service/internal/rbac"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler api.StrictServerInterface, validator *auth.Validator, authorizer *rbac.Authorizer) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middlwr.NewAuth(validator).Handler)
	router.Use(middlwr.NewAuthorization(authorizer).Handler)

	strictHandler := api.NewStrictHandler(handler, nil)

	api.HandlerFromMux(strictHandler, router)

	return router
}