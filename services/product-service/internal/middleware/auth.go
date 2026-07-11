package middleware

import (
	"net/http"
	"strings"

	"github.com/byorty/test-marketplace/services/product-service/internal/auth"
)

type AuthMiddleware struct {
	jwt *auth.Manager
}

func NewAuth(jwt *auth.Manager) *AuthMiddleware {
	return &AuthMiddleware{
		jwt: jwt,
	}
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "missing autorization header", http.StatusUnauthorized)
			return 
		}

		const prefix = "Bearer "

		if !strings.HasPrefix(header, prefix) {
			http.Error(w, "invalid authorization scheme", http.StatusUnauthorized)
			return 
		}

		token := strings.TrimPrefix(header, prefix)

		claims, err := m.jwt.Parse(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return 
		}

		ctx := auth.WithClaims(r.Context(), claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}