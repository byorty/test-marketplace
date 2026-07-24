package http

import (
	"github.com/byorty/test-marketplace/services/order-service/internal/domain/order"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
	"github.com/google/uuid"
)

func toCartItemInput(userID uuid.UUID, req api.AddToCartRequest) *order.CartItem {
	return &order.CartItem{
		ID: uuid.New(),
		ProductID: req.ProductId,
		UserID: userID,
		Quantity: int(req.Quantity),
	}
}

func toCartResponse(cart *order.Cart) api.Cart {
	items := make([]api.CartItem, 0, len(cart.Items))

	for _, item := range cart.Items {
		items = append(items, api.CartItem{
			ProductId: item.ProductID,
			Quantity: int32(item.Quantity),
		})
	}

	return api.Cart{
		Items: items,
		TotalPrice: cart.TotalPrice,
	}
}

func toOrderResponse(o *order.Order) api.Order {
	items := make([]api.OrderItem, 0, len(o.Items))

	for _, item := range o.Items {
		items = append(items, api.OrderItem{
			ProductId: item.ProductID,
			ProductName: item.ProductName,
			ProductPrice: item.ProductPrice,
			Quantity: int32(item.Quantity),
		})
	}

	return api.Order{
		Id: o.ID,
		Status: api.OrderStatus(o.Status),
		TotalPrice: o.Total,
		CreatedAt: o.CreatedAt,
		DeliveryDate: o.DeliveryDate,
		Items: items,
	}
}

func toCreateOrderResp(order *order.Order) api.CreateOrderResponse {
	return api.CreateOrderResponse{
		Id: order.ID,
		Status: api.OrderStatus(order.Status),
	}
}


