package middleware

import (
	"net/http"

	"github.com/byorty/test-marketplace/services/product-service/internal/auth"
	"github.com/byorty/test-marketplace/services/product-service/internal/rbac"
)

type AuthorizationMiddleware struct {
	auth *rbac.Authorizer
}

func NewAuthorization(auth *rbac.Authorizer) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		auth: auth,
	}
}

func (m *AuthorizationMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return 
		}

		path := rbac.NormalizePath(r.URL.Path)

		allowed := m.auth.Authorize(claims.Role, r.Method, path)

		if !allowed {
			http.Error(w, "forbidden", http.StatusForbidden)
		}

		next.ServeHTTP(w, r)
	})
}