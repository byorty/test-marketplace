package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"

	"go.uber.org/zap"
)

type OrderRepository struct {
	db  bun.IDB
	log *zap.Logger
}

func New(db *bun.DB, log *zap.Logger) *OrderRepository {
	return &OrderRepository{
		db:  db,
		log: log.Named("order-repository"),
	}
}

func (r *OrderRepository) AddToCart(ctx context.Context, item *domain.CartItem) error {
	log := r.log.Named("OrderRepository.AddToCart")

	_, err := r.db.NewInsert().
		Model(item).
		Exec(ctx)

	if err != nil {
		log.Error(
			"failed to add item to cart",
			zap.Error(err),
			zap.String("cart_item_id", item.ID.String()),
			zap.String("user_id", item.UserID.String()),
			zap.String("product_id", item.ProductID.String()),
		)

		return err
	}

	return nil
}

func (r *OrderRepository) GetCart(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error) {
	log := r.log.Named("OrderRepository.GetCart")

	items := make([]domain.CartItem, 0)

	err := r.db.NewSelect().
		Model(&items).
		Where("user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		log.Error(
			"failed to query cart",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)

		return nil, err
	}

	if len(items) == 0 {
		return nil, domain.ErrCartEmpty
	}

	return items, nil
}

func (r *OrderRepository) RemoveFromCart(ctx context.Context, userID, productID uuid.UUID) error {
	log := r.log.Named("OrderRepository.RemoveFromCart")

	res, err := r.db.NewDelete().
		Model((*domain.CartItem)(nil)).
		Where("user_id = ?", userID).
		Where("product_id = ?", productID).
		Exec(ctx)

	if err != nil {
		log.Error(
			"failed to remove cart item",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("product_id", productID.String()),
		)

		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrCartItemNotFound
	}

	return nil
}

func (r *OrderRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	log := r.log.Named("OrderRepository.ClearCart")

	_, err := r.db.NewDelete().
		Model((*domain.CartItem)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		log.Error(
			"failed to clear cart",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)

		return err
	}

	return nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, o *domain.Order) error {
	log := r.log.Named("OrderRepository.CreateOrder")

	_, err := r.db.NewInsert().
		Model(o).
		Exec(ctx)

	if err != nil {
		log.Error(
			"failed to create order",
			zap.Error(err),
			zap.String("order_id", o.ID.String()),
			zap.String("user_id", o.UserID.String()),
		)

		return err
	}

	return nil
}

func (r *OrderRepository) CreateOrderItems(ctx context.Context, items []domain.OrderItem) error {
	log := r.log.Named("OrderRepository.CreateOrderItems")

	if len(items) == 0 {
		return nil
	}

	_, err := r.db.NewInsert().
		Model(&items).
		Exec(ctx)

	if err != nil {
		log.Error(
			"failed to create order items",
			zap.Error(err),
			zap.String("order_id", items[0].OrderID.String()),
			zap.Int("items_count", len(items)),
		)

		return err
	}

	return nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	log := r.log.Named("OrderRepository.GetOrderByID")

	order := new(domain.Order)

	err := r.db.NewSelect().
		Model(order).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}

	if err != nil {
		log.Error(
			"failed to get order",
			zap.Error(err),
			zap.String("order_id", id.String()),
		)

		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	log := r.log.Named("OrderRepository.GetOrderItems")

	items := make([]domain.OrderItem, 0)

	err := r.db.NewSelect().
		Model(&items).
		Where("order_id = ?", orderID).
		Scan(ctx)

	if err != nil {
		log.Error(
			"failed to query order items",
			zap.Error(err),
			zap.String("order_id", orderID.String()),
		)

		return nil, err
	}

	return items, nil
}

func (r *OrderRepository) Transaction(ctx context.Context, fn func(repo domain.OrderRepository) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	txRepo := &OrderRepository{
		db:  tx,
		log: r.log,
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}