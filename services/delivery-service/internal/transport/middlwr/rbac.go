package middlwr

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/common/rbac"
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
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		resource, action := permission(r)

		fmt.Printf(
			"RBAC DEBUG: role=%q resource=%q action=%q path=%q\n",
			claims.Role,
			resource,
			action,
			r.URL.Path,
		)

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

func permission(r *http.Request) (rbac.Resource, rbac.Action) {
	switch {
	case r.Method == http.MethodPost &&
		r.URL.Path == "/deliveries/verify-qr":
		return rbac.ResourceDelivery, rbac.ActionVerifyQR

	case r.Method == http.MethodPost &&
		r.URL.Path == "/deliveries":
		return rbac.ResourceDelivery, rbac.ActionCreate

	case r.Method == http.MethodGet &&
		r.URL.Path == "/deliveries":
		return rbac.ResourceDelivery, rbac.ActionView

	case r.Method == http.MethodGet &&
		strings.HasPrefix(r.URL.Path, "/deliveries/") &&
		strings.HasSuffix(r.URL.Path, "/qr"):
		return rbac.ResourceDelivery, rbac.ActionGetQR

	case r.Method == http.MethodGet &&
		strings.HasPrefix(r.URL.Path, "/deliveries/"):
		return rbac.ResourceDelivery, rbac.ActionView

	case r.Method == http.MethodPatch &&
		strings.HasPrefix(r.URL.Path, "/deliveries/") &&
		strings.HasSuffix(r.URL.Path, "/status"):
		return rbac.ResourceDelivery, rbac.ActionUpdateStatus

	case r.Method == http.MethodPatch &&
		strings.HasPrefix(r.URL.Path, "/deliveries/") &&
		strings.HasSuffix(r.URL.Path, "/reschedule"):
		return rbac.ResourceDelivery, rbac.ActionReschedule
	}

	return "", ""
}
