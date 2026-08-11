package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	pr "github.com/byorty/test-marketplace/services/common/client/product"
	"github.com/byorty/test-marketplace/services/order-service/internal/client"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderService struct {
	repo domain.OrderRepository
	log *zap.Logger

	productClient client.Client
}

func New(repo domain.OrderRepository, log *zap.Logger, productClient client.Client) *OrderService {
	return &OrderService{
		repo: repo,
		log: log.Named("order-service"),

		productClient: productClient,
	}
}

func (s *OrderService) AddToCart(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
	start := time.Now()

	if userID == uuid.Nil {
		s.log.Error(
			"add to cart failed",
			zap.Error(ErrInvalidUserID),
		)

		return ErrInvalidUserID
	}

	if item == nil {
		s.log.Error(
			"add to cart failed",
			zap.Error(ErrInvalidInput),
			zap.String("user_id", userID.String()),
		)

		return ErrInvalidInput
	}

	s.log.Info(
		"add to cart started",
		zap.String("user_id", userID.String()),
		zap.String("product_id", item.ProductID.String()),
		zap.Int("quantity", item.Quantity),
	)

	if item.ProductID == uuid.Nil {
		s.log.Error(
			"add to cart failed",
			zap.Error(ErrInvalidProductID),
			zap.String("user_id", userID.String()),
		)

		return ErrInvalidProductID
	}

	if item.Quantity <= 0 {
		s.log.Error(
			"add to cart failed",
			zap.Error(ErrInvalidQuantity),
			zap.String("user_id", userID.String()),
			zap.Int("quantity", item.Quantity),
		)

		return ErrInvalidQuantity
	}

	product, err := s.productClient.GetProduct(ctx, item.ProductID)
	if err != nil {
		if errors.Is(err, pr.ErrProductNotFound) {
			s.log.Warn(
				"product not found",
				zap.String("user_id", userID.String()),
				zap.String("product_id", item.ProductID.String()),
			)

			return ErrProductNotFound
		}

		s.log.Error(
			"get product failed",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("product_id", item.ProductID.String()),
		)

		return fmt.Errorf("get product: %w", err)
	}

	if product == nil {
		s.log.Warn(
			"product not found",
			zap.String("user_id", userID.String()),
			zap.String("product_id", item.ProductID.String()),
		)

		return ErrProductNotFound
	}

	item.UserID = userID

	if err := s.repo.AddToCart(ctx, userID, item); err != nil {
		s.log.Error(
			"add to cart failed",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("product_id", item.ProductID.String()),
		)

		return fmt.Errorf("add to cart: %w", err)
	}

	s.log.Info(
		"add to cart success",
		zap.String("user_id", userID.String()),
		zap.String("product_id", item.ProductID.String()),
		zap.Int("quantity", item.Quantity),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}

func (s *OrderService) GetCart(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	start := time.Now()

	s.log.Debug(
		"get cart started",
		zap.String("user_id", userID.String()),
	)

	if userID == uuid.Nil {
		s.log.Error(
			"invalid user id",
			zap.Error(ErrInvalidUserID),
		)

		return nil, ErrInvalidUserID
	}

	items, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		s.log.Error(
			"get cart failed",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)

		return nil, fmt.Errorf("get cart: %w", err)
	}

	s.log.Debug(
		"get cart success",
		zap.String("user_id", userID.String()),
		zap.Int("items_count", len(items)),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	)

	return &domain.Cart{
		Items:      items,
	}, nil
}

func (s *OrderService) RemoveFromCart(ctx context.Context, userID, productID uuid.UUID) error {
	start := time.Now()

	s.log.Info(
		"remove from cart started",
		zap.String("user_id", userID.String()),
		zap.String("product_id", productID.String()),
	)

	if userID == uuid.Nil {
		s.log.Error("invalid user id", zap.Error(ErrInvalidUserID))
		return ErrInvalidUserID
	}

	if productID == uuid.Nil {
		s.log.Error("invalid product id", zap.Error(ErrInvalidProductID))
		return ErrInvalidProductID
	}

	_, err := s.repo.GetCartItem(ctx, userID, productID)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotInCart) {
			s.log.Warn(
				"product not found in user's cart",
				zap.String("user_id", userID.String()),
				zap.String("product_id", productID.String()),
			)
			return domain.ErrCartItemNotFound 
		}
		return fmt.Errorf("check cart item: %w", err)
	}

	if err := s.repo.RemoveFromCart(ctx, userID, productID); err != nil {
		return fmt.Errorf("remove from cart: %w", err)
	}

	s.log.Info(
		"remove from cart success",
		zap.String("user_id", userID.String()),
		zap.String("product_id", productID.String()),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}

func (s *OrderService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	start := time.Now()

	s.log.Info(
		"clear cart started",
		zap.String("user_id", userID.String()),
	)

	if userID == uuid.Nil {
		s.log.Warn(
			"invalid user id",
			zap.String("user_id", userID.String()),
		)
		return ErrInvalidUserID
	}

	if err := s.repo.ClearCart(ctx, userID); err != nil {
		return fmt.Errorf("clear cart: %w", err)
	}

	s.log.Info(
		"clear cart success",
		zap.String("user_id", userID.String()),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	)

	return nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, userID, orderID uuid.UUID) (*domain.Order, error) {
	s.log.Debug(
		"get order started",
		zap.String("user_id", userID.String()),
		zap.String("order_id", orderID.String()),
	)

	if userID == uuid.Nil {
		s.log.Warn(
			"invalid user id",
			zap.String("user_id", userID.String()),
		)
		return nil, ErrInvalidUserID
	}

	if orderID == uuid.Nil {
		s.log.Warn(
			"invalid order id",
			zap.String("order_id", orderID.String()),
		)
		return nil, ErrInvalidID
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	
	if order.UserID != userID {
		s.log.Warn(
			"order does not belong to user",
			zap.String("user_id", userID.String()),
			zap.String("order_id", orderID.String()),
			zap.String("order_user_id", order.UserID.String()),
		)
		return nil, ErrForbidden 
	}

	s.log.Debug(
		"get order success",
		zap.String("user_id", userID.String()),
		zap.String("order_id", orderID.String()),
	)

	return order, nil
}

func (s *OrderService) GetOrderItems(ctx context.Context, userID, orderID uuid.UUID) ([]domain.OrderItem, error) {
	s.log.Debug(
		"get order items started",
		zap.String("user_id", userID.String()),
		zap.String("order_id", orderID.String()),
	)

	if userID == uuid.Nil {
		s.log.Warn(
			"invalid user id",
			zap.String("user_id", userID.String()),
		)
		return nil, ErrInvalidUserID
	}

	if orderID == uuid.Nil {
		s.log.Warn(
			"invalid order id",
			zap.String("order_id", orderID.String()),
		)
		return nil, ErrInvalidOrderID
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	if order.UserID != userID {
		s.log.Warn(
			"order does not belong to user",
			zap.String("user_id", userID.String()),
			zap.String("order_user_id", order.UserID.String()),
		)
		return nil, ErrForbidden
	}

	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}

	s.log.Debug(
		"get order items success",
		zap.String("order_id", orderID.String()),
		zap.Int("items_count", len(items)),
	)

	return items, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID) (*domain.Order, error) {
	start := time.Now()

	if userID == uuid.Nil {
		s.log.Warn(
			"invalid user id",
			zap.String("user_id", userID.String()),
		)
		return nil, ErrInvalidUserID
	}

	s.log.Info(
		"create order started",
		zap.String("user_id", userID.String()),
	)

	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	if len(cart) == 0 {
		s.log.Warn(
			"cart is empty",
			zap.String("user_id", userID.String()),
		)
		return nil, domain.ErrCartEmpty
	}

	var total int64
	items := make([]domain.OrderItem, len(cart))
	orderID := uuid.New()

	var deliveryDays int

	for _, cartItem := range cart {
		product, err := s.productClient.GetProduct(ctx, cartItem.ProductID)
		if err != nil {
			if errors.Is(err, pr.ErrProductNotFound) {
				s.log.Warn(
					"product not found",
					zap.String("product_id", cartItem.ProductID.String()),
				)
				return nil, ErrProductNotFound
			}

			return nil, fmt.Errorf("get product: %w", err)
		}

		total += product.Price * int64(cartItem.Quantity)

		if product.DeliveryDays > deliveryDays {
			deliveryDays = product.DeliveryDays
		}

		items = append(items, domain.OrderItem{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    product.Id,
			ProductPrice: product.Price,
			Quantity:     cartItem.Quantity,
		})
	}

	order := &domain.Order{
		ID:           orderID,
		UserID:       userID,
		Status:       domain.StatusCreated,
		Total:        total,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(time.Duration(deliveryDays) * 24 * time.Hour),
	}

	if err := s.repo.Transaction(ctx, func(repo domain.OrderRepository) error {
		if err := repo.CreateOrder(ctx, order); err != nil {
			return err
		}

		if err := repo.CreateOrderItems(ctx, items); err != nil {
			return err
		}

		if err := repo.ClearCart(ctx, userID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	s.log.Info(
		"create order success",
		zap.String("order_id", orderID.String()),
		zap.String("user_id", userID.String()),
		zap.Int64("total", total),
		zap.Int("items_count", len(items)),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	)

	return order, nil
}

