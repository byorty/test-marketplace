package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestOrderRepository(t *testing.T) (*OrderRepository, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}

	repo := &OrderRepository{
		db:  mock,
		log: zap.NewNop(),
	}

	return repo, mock
}

func TestOrderRepository_AddToCart(t *testing.T) {
	t.Parallel()

	repo, mock := newTestOrderRepository(t)

	item := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ProductID: uuid.New(),
		Quantity:  2,
	}

	tests := []struct {
		name    string
		prepare func()
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				mock.ExpectExec(`INSERT INTO cart_items`).
					WithArgs(item.ID, item.UserID, item.ProductID, item.Quantity).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: nil,
		},
		{
			name: "db error",
			prepare: func() {
				mock.ExpectExec(`INSERT INTO cart_items`).
					WithArgs(item.ID, item.UserID, item.ProductID, item.Quantity).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			err := repo.AddToCart(context.Background(), item)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepository_GetCart(t *testing.T) {
	t.Parallel()

	repo, mock := newTestOrderRepository(t)

	userID := uuid.New()

	tests := []struct {
		name    string
		prepare func()
		check   func(t *testing.T, items []domain.CartItem)
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				rows := pgxmock.NewRows([]string{
					"id", "user_id", "product_id", "quantity",
				})

				rows.AddRow(
					uuid.New(),
					userID,
					uuid.New(),
					2,
				)

				rows.AddRow(
					uuid.New(),
					userID,
					uuid.New(),
					1,
				)

				mock.ExpectQuery(`SELECT id, user_id, product_id, quantity`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, items []domain.CartItem) {
				require.Len(t, items, 2)
				require.Equal(t, userID, items[0].UserID)
				require.Equal(t, 2, items[0].Quantity)
				require.Equal(t, userID, items[1].UserID)
				require.Equal(t, 1, items[1].Quantity)
			},
			wantErr: nil,
		},
		{
			name: "empty cart",
			prepare: func() {
				rows := pgxmock.NewRows([]string{
					"id", "user_id", "product_id", "quantity",
				})

				mock.ExpectQuery(`SELECT id, user_id, product_id, quantity`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, items []domain.CartItem) {
				require.Nil(t, items)
			},
			wantErr: domain.ErrCartEmpty,
		},
		{
			name: "query error",
			prepare: func() {
				mock.ExpectQuery(`SELECT id, user_id, product_id, quantity`).
					WithArgs(userID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "scan error",
			prepare: func() {
				rows := pgxmock.NewRows([]string{
					"id", "user_id", "product_id", "quantity",
				})

				rows.AddRow(
					"not-a-uuid", 
					userID,
					uuid.New(),
					2,
				)

				mock.ExpectQuery(`SELECT id, user_id, product_id, quantity`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantErr: errors.New("Scan"), 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			items, err := repo.GetCart(context.Background(), userID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.check != nil {
				tt.check(t, items)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepository_RemoveFromCart(t *testing.T) {
	t.Parallel()

	repo, mock := newTestOrderRepository(t)

	userID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name    string
		prepare func()
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				mock.ExpectExec(`DELETE FROM cart_items`).
					WithArgs(userID, productID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: nil,
		},
		{
			name: "item not found",
			prepare: func() {
				mock.ExpectExec(`DELETE FROM cart_items`).
					WithArgs(userID, productID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantErr: domain.ErrCartItemNotFound,
		},
		{
			name: "db error",
			prepare: func() {
				mock.ExpectExec(`DELETE FROM cart_items`).
					WithArgs(userID, productID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			err := repo.RemoveFromCart(context.Background(), userID, productID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepository_ClearCart(t *testing.T) {
	t.Parallel()

	repo, mock := newTestOrderRepository(t)

	userID := uuid.New()

	tests := []struct {
		name    string
		prepare func()
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				mock.ExpectExec(`DELETE FROM cart_items`).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 3))
			},
			wantErr: nil,
		},
		{
			name: "db error",
			prepare: func() {
				mock.ExpectExec(`DELETE FROM cart_items`).
					WithArgs(userID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			err := repo.ClearCart(context.Background(), userID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepository_CreateOrder(t *testing.T) {
	t.Parallel()

	repo, mock := newTestOrderRepository(t)

	now := time.Now()
	deliveryDate := now.Add(5 * 24 * time.Hour)

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       "pending",
		Total:        150000,
		CreatedAt:    now,
		DeliveryDate: deliveryDate,
	}

	tests := []struct {
		name    string
		prepare func()
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				mock.ExpectExec(`INSERT INTO orders`).
					WithArgs(
						order.ID,
						order.UserID,
						order.Status,
						order.Total,
						order.CreatedAt,
						order.DeliveryDate,
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: nil,
		},
		{
			name: "db error",
			prepare: func() {
				mock.ExpectExec(`INSERT INTO orders`).
					WithArgs(
						order.ID,
						order.UserID,
						order.Status,
						order.Total,
						order.CreatedAt,
						order.DeliveryDate,
					).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			err := repo.CreateOrder(context.Background(), order)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepository_CreateOrderItems(t *testing.T) {
	t.Parallel()

	repo, mock := newTestOrderRepository(t)

	orderID := uuid.New()

	items := []domain.OrderItem{
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 50000,
			Quantity:     1,
		},
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 100000,
			Quantity:     2,
		},
	}

	tests := []struct {
		name    string
		prepare func()
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				mock.ExpectExec(`INSERT INTO order_items`).
					WithArgs(
						items[0].ID,
						items[0].OrderID,
						items[0].ProductID,
						items[0].ProductPrice,
						items[0].Quantity,
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectExec(`INSERT INTO order_items`).
					WithArgs(
						items[1].ID,
						items[1].OrderID,
						items[1].ProductID,
						items[1].ProductPrice,
						items[1].Quantity,
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: nil,
		},
		{
			name: "batch error on first item",
			prepare: func() {
				mock.ExpectExec(`INSERT INTO order_items`).
					WithArgs(
						items[0].ID,
						items[0].OrderID,
						items[0].ProductID,
						items[0].ProductPrice,
						items[0].Quantity,
					).
					WillReturnError(errors.New("batch error"))
			},
			wantErr: errors.New("batch error"),
		},
		{
			name: "batch error on second item",
			prepare: func() {
				mock.ExpectExec(`INSERT INTO order_items`).
					WithArgs(
						items[0].ID,
						items[0].OrderID,
						items[0].ProductID,
						items[0].ProductPrice,
						items[0].Quantity,
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectExec(`INSERT INTO order_items`).
					WithArgs(
						items[1].ID,
						items[1].OrderID,
						items[1].ProductID,
						items[1].ProductPrice,
						items[1].Quantity,
					).
					WillReturnError(errors.New("batch error"))
			},
			wantErr: errors.New("batch error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			err := repo.CreateOrderItems(context.Background(), items)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepository_GetOrderByID(t *testing.T) {
	t.Parallel()

	repo, mock := newTestOrderRepository(t)

	orderID := uuid.New()
	now := time.Now()
	deliveryDate := now.Add(5 * 24 * time.Hour)

	tests := []struct {
		name    string
		prepare func()
		check   func(t *testing.T, order *domain.Order)
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				rows := pgxmock.NewRows([]string{
					"id", "user_id", "status", "total_price", "created_at", "delivery_date",
				})

				rows.AddRow(
					orderID,
					uuid.New(),
					domain.Status("pending"),
					int64(150000),
					now,
					deliveryDate,
				)

				mock.ExpectQuery(`SELECT id, user_id, status, total_price, created_at, delivery_date`).
					WithArgs(orderID).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, order *domain.Order) {
				require.NotNil(t, order)
				require.Equal(t, orderID, order.ID)
				require.Equal(t, domain.Status("pending"), order.Status)
				require.Equal(t, int64(150000), order.Total)
			},
			wantErr: nil,
		},
		{
			name: "order not found",
			prepare: func() {
				mock.ExpectQuery(`SELECT id, user_id, status, total_price, created_at, delivery_date`).
					WithArgs(orderID).
					WillReturnError(pgx.ErrNoRows)
			},
			check: func(t *testing.T, order *domain.Order) {
				require.Nil(t, order)
			},
			wantErr: domain.ErrOrderNotFound,
		},
		{
			name: "db error",
			prepare: func() {
				mock.ExpectQuery(`SELECT id, user_id, status, total_price, created_at, delivery_date`).
					WithArgs(orderID).
					WillReturnError(errors.New("db error"))
			},
			check: func(t *testing.T, order *domain.Order) {
				require.Nil(t, order)
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			order, err := repo.GetOrderByID(context.Background(), orderID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.check != nil {
				tt.check(t, order)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepository_GetOrderItems(t *testing.T) {
	t.Parallel()

	repo, mock := newTestOrderRepository(t)

	orderID := uuid.New()

	tests := []struct {
		name    string
		prepare func()
		check   func(t *testing.T, items []domain.OrderItem)
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				rows := pgxmock.NewRows([]string{
					"id", "order_id", "product_id", "product_price", "quantity",
				})

				rows.AddRow(
					uuid.New(),
					orderID,
					uuid.New(),
					int64(50000),
					1,
				)

				rows.AddRow(
					uuid.New(),
					orderID,
					uuid.New(),
					int64(100000),
					2,
				)

				mock.ExpectQuery(`SELECT id, order_id, product_id, product_price, quantity`).
					WithArgs(orderID).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, items []domain.OrderItem) {
				require.Len(t, items, 2)
				require.Equal(t, orderID, items[0].OrderID)
				require.Equal(t, int64(50000), items[0].ProductPrice)
				require.Equal(t, 1, items[0].Quantity)
				require.Equal(t, orderID, items[1].OrderID)
				require.Equal(t, int64(100000), items[1].ProductPrice)
				require.Equal(t, 2, items[1].Quantity)
			},
			wantErr: nil,
		},
		{
			name: "empty items",
			prepare: func() {
				rows := pgxmock.NewRows([]string{
					"id", "order_id", "product_id", "product_price", "quantity",
				})

				mock.ExpectQuery(`SELECT id, order_id, product_id, product_price, quantity`).
					WithArgs(orderID).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, items []domain.OrderItem) {
				require.Empty(t, items)
			},
			wantErr: nil,
		},
		{
			name: "query error",
			prepare: func() {
				mock.ExpectQuery(`SELECT id, order_id, product_id, product_price, quantity`).
					WithArgs(orderID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			items, err := repo.GetOrderItems(context.Background(), orderID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.check != nil {
				tt.check(t, items)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

