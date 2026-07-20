package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/byorty/test-marketplace/services/order-service/internal/client/product"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain/order"
	"github.com/google/uuid"
)

type Service struct {
	repo order.Repository
	log *slog.Logger

	productClient product.Client
}

func New(repo order.Repository, log *slog.Logger, productClient product.Client) *Service {
	return &Service{
		repo: repo,
		log: log.With(slog.String("layer", "service")),

		productClient: productClient,
	}
}

func logError(log *slog.Logger, op string, err error) {
	log.Error("operation failed", slog.String("op", op), slog.Any("error", err))
}

func (s *Service) AddToCart(ctx context.Context, item *order.CartItem) error {
	const op = "Service.AddToCart"

	start := time.Now()

	s.log.Info("add to cart started", "op", op,
	"user_id", item.UserID,
	"product_id", item.ProductID,
	"quantiy", item.Quantity)

	if item == nil {
		logError(s.log, op, ErrInvalidInput)
		return ErrInvalidInput
	}

	if item.UserID == uuid.Nil {
		logError(s.log, op, ErrInvalidUserID)
		return ErrInvalidUserID
	}

	if item.ProductID == uuid.Nil {
		logError(s.log, op, ErrInvaliProductdID)
		return ErrInvaliProductdID
	}

	if item.Quantity <= 0 {
		logError(s.log, op, ErrInvalidQuantity)
		return ErrInvalidQuantity
	}

	if err := s.repo.AddToCart(ctx, item); err != nil {
		logError(s.log, op, err)
		return fmt.Errorf("%s: add to cart: %w", op, err)
	}

	s.log.Info("add to cart success", "op", op, "item", item, "duration_ms", time.Since(start).Milliseconds())

	return nil
}

func (s *Service) GetCart(ctx context.Context, userID uuid.UUID) ([]order.CartItem, error) {
	const op = "Service.GetCart"

	s.log.Debug("get cart", "op", op, "user_id", userID)

	if userID == uuid.Nil {
		logError(s.log, op, ErrInvalidUserID)
		return nil, ErrInvalidUserID 
	}

	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		logError(s.log, op, err)
		return nil, fmt.Errorf("%s: get cart: %w", op, err)
	}

	return cart, nil
}

func (s *Service) RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error {
	const op = "Service.RemoveFromCart"

	start := time.Now()

	s.log.Info("remove from cart started", "op", op, "user_id", userID, "product_id", productID)

	if userID == uuid.Nil {
		logError(s.log, op, ErrInvalidUserID)
		return ErrInvalidUserID
	}

	if productID == uuid.Nil {
		logError(s.log, op, ErrInvaliProductdID)
		return ErrInvaliProductdID
	}

	if err := s.repo.RemoveFromCart(ctx, userID, productID); err != nil {
		logError(s.log, op, err)
		return fmt.Errorf("%s: remove from cart: %w", op, err)
	}

	s.log.Info("remove from cart success", 
	"op", op, "user_id", userID, 
	"product_id", productID, 
	"duration_ms", time.Since(start).Milliseconds())

	return nil
}

func (s *Service) ClearCart(ctx context.Context, userID uuid.UUID) error {
	const op = "Service.ClearCart"

	start := time.Now()

	s.log.Info("clear cart started", "op", op, "user_id", userID)

	if userID == uuid.Nil {
		logError(s.log, op, ErrInvalidUserID)
		return ErrInvalidUserID
	}

	if err := s.repo.ClearCart(ctx, userID); err != nil {
		logError(s.log, op, err)
		return fmt.Errorf("%s: clear cart: %w", op, err)
	}

	s.log.Info("clear cart success", "op", op, 
	"user_id", userID, "duration_ms", time.Since(start).Milliseconds())

	return nil
}	

func (s *Service) GetOrderByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	const op = "Service.GetOrderByID"

	s.log.Debug("get order by id", "op", op, "id", id)

	if id == uuid.Nil {
		logError(s.log, op, ErrInvalidID)
		return nil, ErrInvalidID
	}

	o, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		logError(s.log, op, err)
		return nil, fmt.Errorf("%s: get order by id: %w", op, err)
	}

	return o, nil
}

func (s *Service) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]order.OrderItem, error) {
	const op = "Service.GetOrderItems"

	s.log.Debug("get order items", "op", op, "order_id", orderID)

	if orderID == uuid.Nil {
		logError(s.log, op, ErrInvalidOrderID)
		return nil, ErrInvalidOrderID
	}

	o, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		logError(s.log, op, err)
		return nil, fmt.Errorf("%s: get order items: %w", op, err)
	}

	return o, nil
}

func (s *Service) CreateOrder(ctx context.Context, userID uuid.UUID) (*order.Order, error) {
	const op = "Service.CreateOrder"

	start := time.Now()

	if userID == uuid.Nil {
		logError(s.log, op, ErrInvalidUserID)
		return nil, ErrInvalidUserID
	}

	s.log.Info("create order started", "op", op, "user_id", userID)

	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		logError(s.log, op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(cart) == 0 {
		logError(s.log, op, order.ErrCartEmpty)
		return nil, order.ErrCartEmpty
	}

	total := int64(0)

	items := make([]order.OrderItem, 0, len(cart))

	orderID := uuid.New()

	deliveryDays := 0

	for _, cartItem := range cart {

		product, err := s.productClient.GetByID(ctx, cartItem.ProductID)

		if err != nil {
			logError(s.log, op, err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		
		total += product.Price * int64(cartItem.Quantity)

		if product.DeliveryDays > deliveryDays {
			deliveryDays = product.DeliveryDays
		}

		items = append(items, order.OrderItem{
			ID: uuid.New(),
			OrderID: orderID,
			ProductID: product.ID,
			ProductName: product.Name,
			ProductPrice: product.Price,
			Quantity: cartItem.Quantity,
		})
	}

	o := &order.Order{
		ID: orderID,
		UserID: userID,
		Status: order.StatusCreated,
		Total: total,
		CreatedAt: time.Now(),
		DeliveryDate: time.Now().Add(time.Duration(deliveryDays) * 24 * time.Hour),
	}

	err = s.repo.Transaction(ctx, func(repo order.Repository) error {
		if err := s.repo.CreateOrder(ctx, o); err != nil {
			return err
		}

		if err := s.repo.CreateOrderItems(ctx, items); err != nil {
			return err
		}

		if err := s.repo.ClearCart(ctx, userID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logError(s.log, op, err)
		return nil, fmt.Errorf("%s: create order: %w", op, err)
	}

	s.log.Info("create order success", "op", op, 
	"order_id", orderID, "user_id", userID,
	"duration_ms", time.Since(start).Milliseconds())

	return o, nil
}