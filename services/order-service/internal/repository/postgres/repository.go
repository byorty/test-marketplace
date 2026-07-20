package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain/order"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) AddToCart(ctx context.Context, item *order.CartItem) error {
	const op = "Repository.AddToCart"

	model := cartItemToModel(item)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *Repository) GetCart(ctx context.Context, userID uuid.UUID) ([]order.CartItem, error) {
	const op = "Repository.GetCart"
	var models []CartItemModel

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	
	if len(models) == 0 {
		return nil, order.ErrCartEmpty
	}

	items := make([]order.CartItem, 0, len(models))
	for _, m := range models {
		items = append(items, cartItemToDomain(m))
	}

	return items, nil
} 

func (r *Repository) RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error {
	const op = "Repository.RemoveFromCart"

	result := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).Delete(&CartItemModel{})

	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		return order.ErrCartItemNotFound
	}

	return nil
}

func (r *Repository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	const op = "Repository.ClearCart"

	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&CartItemModel{})

	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	return nil
}

func (r *Repository) CreateOrder(ctx context.Context, order *order.Order) error {
	const op = "Repository.CreateOrder"

	model := orderToModel(order)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) CreateOrderItems(ctx context.Context, items []*order.OrderItem) error {
	const op = "Repository.CreateOrderItems"

	models := make([]OrderItemModel, 0, len(items))
	for _, item := range items {
		models = append(models, orderItemToModel(item))
	}

	err := r.db.WithContext(ctx).Create(&models).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) GetOrderByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	const op = "Repository.GetOrderByID"
	var model OrderModel

	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, order.ErrOrderNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return orderToDomain(model), nil
}

func (r *Repository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]order.OrderItem, error) {
	const op = "Repository.GetOrderItems"
	var models []OrderItemModel

	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	items := make([]order.OrderItem, 0, len(models))
	for _, m := range models {
		 items = append(items, orderItemToDomain(m))
	}

	return items, nil
}