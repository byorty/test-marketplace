package middlwr

import (
	"context"
	"net/http"
	"strings"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	"github.com/byorty/test-marketplace/services/order-service/internal/rbac"
)

type Authorization struct {
	authorizer *rbac.Authorizer
}

func NewAuthorization(a *rbac.Authorizer) *Authorization {
	return &Authorization{
		authorizer: a,
	}
}

func (a *Authorization) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return 
		}

		resource, action := permission(r)

		if err := a.authorizer.Authorize(claims.Role, resource, action); err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return 
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), struct{}{}, claims)))
	})
}

func permission(r *http.Request) (string, string) {
	switch {

	case r.Method == http.MethodPost &&
		r.URL.Path == "/cart/items":

		return rbac.ResourceCart, rbac.ActionAdd

	case r.Method == http.MethodDelete &&
		strings.HasPrefix(r.URL.Path, "/cart/items"):

		return rbac.ResourceCart, rbac.ActionRemove

	case r.Method == http.MethodGet &&
		r.URL.Path == "/cart":

		return rbac.ResourceCart, rbac.ActionView

	case r.Method == http.MethodPost && 
		r.URL.Path == "/order":

		return rbac.ResourceOrder, rbac.ActionCreate

	case r.Method == http.MethodGet && 
		strings.HasPrefix(r.URL.Path, "/orders/"):

		return rbac.ResourceOrder, rbac.ActionView
	}

	return "", ""
}