package postgres

import "github.com/byorty/test-marketplace/services/order-service/internal/domain/order"

func orderToModel(d *order.Order) OrderModel {
	return OrderModel{
		ID: d.ID,
		UserID: d.UserID,
		Status: d.Status,
		Total: d.Total,
		CreatedAt: d.CreatedAt,
		DeliveryDate: d.DeliveryDate,
	}
}

func orderToDomain(p OrderModel) *order.Order {
	return &order.Order{
		ID: p.ID,
		UserID: p.UserID,
		Status: p.Status,
		Total: p.Total,
		CreatedAt: p.CreatedAt,
		DeliveryDate: p.DeliveryDate,
	}
}

func orderItemToModel(d order.OrderItem) OrderItemModel {
	return OrderItemModel{
		ID: d.ID,
		OrderID: d.OrderID,
		ProductID: d.ProductID,
		ProductName: d.ProductName,
		ProductPrice: d.ProductPrice,
		Quantity: d.Quantity,
	}
}

func orderItemToDomain(p OrderItemModel) order.OrderItem {
	return order.OrderItem{
		ID: p.ID,
		OrderID: p.OrderID,
		ProductID: p.ProductID,
		ProductName: p.ProductName,
		ProductPrice: p.ProductPrice,
		Quantity: p.Quantity,
	}
}

func cartItemToModel(d *order.CartItem) CartItemModel {
	return CartItemModel{
		ID: d.ID,
		UserID: d.UserID,
		ProductID: d.ProductID,
		Quantity: d.Quantity,
	}
}
 
func cartItemToDomain(p CartItemModel) order.CartItem {
	return order.CartItem{
		ID: p.ID,
		UserID: p.UserID,
		ProductID: p.ProductID,
		Quantity: p.Quantity,
	}
}