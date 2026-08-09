package service

import (
	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
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