package repository

import (
	db "github.com/byorty/test-marketplace/services/common/db"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
)

func toDBOrder(o *domain.Order) *db.Order {
	if o == nil {
		return nil
	}

	items := make([]db.OrderItem, 0, len(o.Items))

	for _, item := range o.Items {
		items = append(items, *toDBOrderItem(&item))
	}

	return &db.Order{
		ID:           o.ID,
		UserID:       o.UserID,
		Status:       string(o.Status),
		TotalPrice:   o.TotalPrice,
		CreatedAt:    o.CreatedAt,
		DeliveryDate: o.DeliveryDate,
	}
}

func toDomainOrder(o *db.Order) *domain.Order {
	if o == nil {
		return nil
	}

	items := make([]domain.OrderItem, 0, len(o.Items))

	for i := range o.Items {
		items = append(
			items,
			*toDomainOrderItem(&o.Items[i]),
		)
	}

	return &domain.Order{
		ID:           o.ID,
		UserID:       o.UserID,
		Status:       domain.Status(o.Status),
		TotalPrice:   o.TotalPrice,
		CreatedAt:    o.CreatedAt,
		DeliveryDate: o.DeliveryDate,
		Items:        items,
	}
}

func toDBOrderItem(item *domain.OrderItem) *db.OrderItem {
	if item == nil {
		return nil
	}

	return &db.OrderItem{
		ID:           item.ID,
		OrderID:      item.OrderID,
		ProductID:    item.ProductID,
		ProductPrice: item.ProductPrice,
		Quantity:     item.Quantity,
	}
}

func toDomainOrderItem(item *db.OrderItem) *domain.OrderItem {
	if item == nil {
		return nil
	}

	return &domain.OrderItem{
		ID:           item.ID,
		OrderID:      item.OrderID,
		ProductID:    item.ProductID,
		ProductPrice: item.ProductPrice,
		Quantity:     item.Quantity,
	}
}

func toDBCartItem(item *domain.CartItem) *db.CartItem {
	if item == nil {
		return nil
	}

	return &db.CartItem{
		ID:        item.ID,
		UserID:    item.UserID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
	}
}

func toDomainCartItem(item *db.CartItem) *domain.CartItem {
	if item == nil {
		return nil
	}

	return &domain.CartItem{
		ID:        item.ID,
		UserID:    item.UserID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
	}
}

