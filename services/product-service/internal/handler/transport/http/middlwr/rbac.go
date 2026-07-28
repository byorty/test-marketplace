package middlwr

import (
	"net/http"
	"strings"

	"github.com/byorty/test-marketplace/services/product-service/internal/auth"
	"github.com/byorty/test-marketplace/services/product-service/internal/rbac"
)

type Authorization struct {
	authorizer *rbac.Authorizer
}

func NewAuthorization(a *rbac.Authorizer) *Authorization {
	return &Authorization{
		authorizer: a,
	}
}

func (m *Authorization) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		resource, action := permission(r)

		if err := m.authorizer.Authorize(
			claims.Role,
			resource,
			action,
		); err != nil {

			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func permission(r *http.Request) (string, string) {
	switch {

	case r.Method == http.MethodPost &&
		r.URL.Path == "/products":

		return rbac.ResourceProduct, rbac.ActionCreate

	case r.Method == http.MethodPatch &&
		strings.HasPrefix(r.URL.Path, "/products/"):

		return rbac.ResourceProduct, rbac.ActionUpdate

	case r.Method == http.MethodDelete &&
		strings.HasPrefix(r.URL.Path, "/products/"):

		return rbac.ResourceProduct, rbac.ActionDelete

	case r.Method == http.MethodGet &&
		r.URL.Path == "/products":

		return rbac.ResourceProduct, rbac.ActionView

	case r.Method == http.MethodGet &&
		strings.HasPrefix(r.URL.Path, "/products/"):

		return rbac.ResourceProduct, rbac.ActionView
	}

	return "", ""
}