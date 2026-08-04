package repository

import (
	"context"
	"errors"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type OrderRepository struct {
	pool *pgxpool.Pool
	db domain.DBTX
	log *zap.Logger
}

func New(pool *pgxpool.Pool, db domain.DBTX, log *zap.Logger) *OrderRepository {
	return &OrderRepository{
		pool: pool,
		db: db,
		log: log.Named("order-repository"),
	}
}

func (r *OrderRepository) AddToCart(ctx context.Context, item *domain.CartItem) error {
	log := r.log.Named("OrderRepository.AddToCart")

	_, err := r.db.Exec(
		ctx,
		`INSERT INTO cart_items (id, user_id, product_id, quantity)
		VALUES ($1, $2, $3, $4)`,
		item.ID,
		item.UserID,
		item.ProductID,
		item.Quantity,
	)

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

	rows, err := r.db.Query(
		ctx,
		`SELECT id, user_id, product_id, quantity
		FROM cart_items
		WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		log.Error(
			"failed to query cart",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)

		return nil, err
	}
	defer rows.Close()

	var items []domain.CartItem

	for rows.Next() {
		var item domain.CartItem

		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ProductID,
			&item.Quantity,
		); err != nil {

			log.Error(
				"failed to scan cart item",
				zap.Error(err),
				zap.String("user_id", userID.String()),
			)

			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		log.Error(
			"failed to iterate cart rows",
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

	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2`,
		userID,
		productID,
	)
	if err != nil {
		log.Error(
			"failed to remove cart item",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("product_id", productID.String()),
		)

		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrCartItemNotFound
	}

	return nil
}

func (r *OrderRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	log := r.log.Named("OrderRepository.ClearCart")

	_, err := r.db.Exec(
		ctx,
		`DELETE FROM cart_items WHERE user_id = $1`,
		userID,
	)
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

	_, err := r.db.Exec(
		ctx,
		`INSERT INTO orders (id, user_id, status, total_price, created_at, delivery_date)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		o.ID,
		o.UserID,
		o.Status,
		o.Total,
		o.CreatedAt,
		o.DeliveryDate,
	)
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

    for _, item := range items {
        _, err := r.db.Exec(
            ctx,
            `INSERT INTO order_items (id, order_id, product_id, product_price, quantity)
            VALUES ($1, $2, $3, $4, $5)`,
            item.ID,
            item.OrderID,
            item.ProductID,
            item.ProductPrice,
            item.Quantity,
        )
        if err != nil {
            log.Error(
                "failed to create order items",
                zap.Error(err),
                zap.String("order_id", items[0].OrderID.String()),
                zap.Int("items_count", len(items)),
            )
            return err
        }
    }

    return nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	log := r.log.Named("OrderRepository.GetOrderByID")

	var o domain.Order

	err := r.db.QueryRow(
		ctx,
		`SELECT id, user_id, status, total_price, created_at, delivery_date
		FROM orders
		WHERE id = $1`,
		id,
	).Scan(
		&o.ID,
		&o.UserID,
		&o.Status,
		&o.Total,
		&o.CreatedAt,
		&o.DeliveryDate,
	)

	if errors.Is(err, pgx.ErrNoRows) {
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

	return &o, nil
}

func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	log := r.log.Named("OrderRepository.GetOrderItems")

	rows, err := r.db.Query(
		ctx,
		`SELECT id, order_id, product_id, product_price, quantity
		FROM order_items
		WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		log.Error(
			"failed to query order items",
			zap.Error(err),
			zap.String("order_id", orderID.String()),
		)

		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem

	for rows.Next() {
		var item domain.OrderItem

		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductPrice,
			&item.Quantity,
		); err != nil {

			log.Error(
				"failed to scan order item",
				zap.Error(err),
				zap.String("order_id", orderID.String()),
			)

			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		log.Error(
			"failed to iterate order items",
			zap.Error(err),
			zap.String("order_id", orderID.String()),
		)

		return nil, err
	}

	return items, nil
}

func (r *OrderRepository) Transaction(ctx context.Context, fn func(repo domain.OrderRepository) error) error {

    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return err
    }

    defer func() {
        _ = tx.Rollback(ctx)
    }()

    txRepo := &OrderRepository{
        db: tx,
        pool: r.pool,
    }

    if err := fn(txRepo); err != nil {
        return err
    }

    return tx.Commit(ctx)
}