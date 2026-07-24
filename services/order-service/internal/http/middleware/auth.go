package middleware

import (
	"net/http"
	"strings"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
)

type Auth struct {
	validator *auth.Validator
}

func New(validator *auth.Validator) *Auth {
	return &Auth{
		validator: validator,
	}
}

func (a *Auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "missing autorization header", http.StatusUnauthorized)
			return 
		}

		const prefix = "Bearer "

		if !strings.HasPrefix(header, prefix) {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return 
		}

		token := strings.TrimPrefix(header, prefix)

		claims, err := a.validator.Parse(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return 
		}

		ctx := auth.ContextWithClaims(r.Context(), claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}