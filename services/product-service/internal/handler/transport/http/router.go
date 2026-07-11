package http

import (
	"net/http"

	"github.com/byorty/test-marketplace/services/product-service/internal/auth"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	middlwr "github.com/byorty/test-marketplace/services/product-service/internal/middleware"
	"github.com/byorty/test-marketplace/services/product-service/internal/rbac"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler api.StrictServerInterface, jwtManager *auth.Manager, authorizer *rbac.Authorizer) http.Handler {
	router := chi.NewRouter()

	authMiddleware := middlwr.NewAuth(jwtManager)
	authorizationMiddleware := middlwr.NewAuthorization(authorizer)

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(authMiddleware.Handler)
	router.Use(authorizationMiddleware.Handler)

	strictHandler := api.NewStrictHandler(handler, nil)

	api.HandlerFromMux(strictHandler, router)

	return router
}