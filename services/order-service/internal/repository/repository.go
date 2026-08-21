package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	db "github.com/byorty/test-marketplace/services/common/db"
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

func (r *OrderRepository) AddToCart(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
	log := r.log.Named("OrderRepository.AddToCart")

	dbItem := toDBCartItem(item)

	_, err := r.db.NewInsert().
		Model(dbItem).
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

	var dbItems []db.CartItem

	err := r.db.NewSelect().
		Model(&dbItems).
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

	if len(dbItems) == 0 {
		return nil, domain.ErrCartEmpty
	}

	items := make([]domain.CartItem, 0, len(dbItems))

	for i := range dbItems {
		items = append(items, *toDomainCartItem(&dbItems[i]))
	}

	return items, nil
}

func (r *OrderRepository) GetCartItem(ctx context.Context, userID, productID uuid.UUID) (*domain.CartItem, error) {
	var dbItem db.CartItem

	err := r.db.NewSelect().
		Model(&dbItem).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCartItemNotFound
		}

		return nil, fmt.Errorf("get cart item: %w", err)
	}

	return toDomainCartItem(&dbItem), nil
}

func (r *OrderRepository) RemoveFromCart(ctx context.Context, userID uuid.UUID, cartItemID uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*db.CartItem)(nil)).
		Where("id = ?", cartItemID).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("delete cart item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrCartItemNotFound
	}

	return nil
}

func (r *OrderRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	log := r.log.Named("OrderRepository.ClearCart")

	_, err := r.db.NewDelete().
		Model((*db.CartItem)(nil)).
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

func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	dbOrder := toDBOrder(order)

	_, err := r.db.NewInsert().
		Model(dbOrder).
		Exec(ctx)

	if err != nil {
		return err
	}

	if len(order.Items) > 0 {
		for i := range order.Items {
			order.Items[i].OrderID = order.ID
		}

		dbItems := make([]db.OrderItem, 0, len(order.Items))

		for i := range order.Items {
			dbItems = append(
				dbItems,
				*toDBOrderItem(&order.Items[i]),
			)
		}

		_, err = r.db.NewInsert().
			Model(&dbItems).
			Exec(ctx)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *OrderRepository) CreateOrderItems(ctx context.Context, items []domain.OrderItem) error {
	log := r.log.Named("OrderRepository.CreateOrderItems")

	if len(items) == 0 {
		return nil
	}

	dbItems := make([]*db.OrderItem, 0, len(items))

	for i := range items {
		dbItems = append(dbItems, toDBOrderItem(&items[i]))
	}

	_, err := r.db.NewInsert().
		Model(&dbItems).
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

	dbOrder := new(db.Order)

	err := r.db.NewSelect().
		Model(dbOrder).
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

	return toDomainOrder(dbOrder), nil
}

func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	log := r.log.Named("OrderRepository.GetOrderItems")

	dbItems := make([]db.OrderItem, 0)

	err := r.db.NewSelect().
		Model(&dbItems).
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

	items := make([]domain.OrderItem, 0, len(dbItems))

	for i := range dbItems {
		items = append(items, *toDomainOrderItem(&dbItems[i]))
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