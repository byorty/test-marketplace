package service

import (
	"github.com/byorty/test-marketplace/services/auth"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/byorty/test-marketplace/services/rbac"
)

func CanAccessOrder(claims *auth.Claims, o *domain.Order) error {

	if claims.Role == "employee" {
		return nil
	}

	if claims.UserID == o.UserID {
		return nil
	}

	return rbac.ErrAccessDenied
}