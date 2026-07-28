package rbac

import (
	"errors"

	"github.com/byorty/test-marketplace/services/order-service/internal/auth"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain/order"
)

var ErrAccessDenied = errors.New("access denied")

func (a *Authorizer) CanViewOrder(claims *auth.Claims, o *order.Order) error {

	if claims.Role == "employee" {
		return nil
	}

	if claims.Role == "customer" &&
		claims.UserID != o.UserID {
		return nil
	}

	return ErrAccessDenied
}

func (a *Authorizer) CanModifyOrder(claims *auth.Claims, o *order.Order) error {

	if claims.Role == "employee" {
		return nil
	}

	if claims.Role == "customer" &&
		claims.UserID == o.UserID {
			return nil
		}

	return ErrAccessDenied
}

func IsAccessDenied(err error) bool {
	return errors.Is(err, ErrAccessDenied)
}