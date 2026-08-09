package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/zap"
)

func newTestDB(t *testing.T) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open(
		"pgx",
		"postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable",
	)
	require.NoError(t, err)

	db := bun.NewDB(
		sqldb,
		pgdialect.New(),
	)

	err = db.Ping()
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func newOrderTestRepository(t *testing.T) (*OrderRepository, *bun.DB) {
	t.Helper()

	db := newTestDB(t)

	repo := New(
		db,
		zap.NewNop(),
	)

	return repo, db
}

func TestOrderRepository_AddToCart(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()

	item := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: uuid.New(),
		Quantity:  2,
	}

	err := repo.AddToCart(ctx, item)

	require.NoError(t, err)

	var actual domain.CartItem

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", item.ID).
		Scan(ctx)

	require.NoError(t, err)

	require.Equal(t, item.ID, actual.ID)
	require.Equal(t, item.UserID, actual.UserID)
	require.Equal(t, item.ProductID, actual.ProductID)
	require.Equal(t, item.Quantity, actual.Quantity)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.CartItem)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
	})
}

func TestOrderRepository_GetCart(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()

	items := []domain.CartItem{
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: uuid.New(),
			Quantity:  2,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: uuid.New(),
			Quantity:  5,
		},
	}

	_, err := db.NewInsert().
		Model(&items).
		Exec(ctx)

	require.NoError(t, err)

	result, err := repo.GetCart(ctx, userID)

	require.NoError(t, err)
	require.Len(t, result, 2)

	require.Equal(t, items[0].ID, result[0].ID)
	require.Equal(t, items[0].UserID, result[0].UserID)
	require.Equal(t, items[0].ProductID, result[0].ProductID)
	require.Equal(t, items[0].Quantity, result[0].Quantity)

	require.Equal(t, items[1].ID, result[1].ID)
	require.Equal(t, items[1].UserID, result[1].UserID)
	require.Equal(t, items[1].ProductID, result[1].ProductID)
	require.Equal(t, items[1].Quantity, result[1].Quantity)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.CartItem)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
	})
}

func TestOrderRepository_GetCart_Empty(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()

	items, err := repo.GetCart(ctx, userID)

	require.Nil(t, items)
	require.ErrorIs(t, err, domain.ErrCartEmpty)

	db.Close()
}

func TestOrderRepository_RemoveFromCart(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()
	productID := uuid.New()

	itemToRemove := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		Quantity:  2,
	}

	itemToKeep := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: uuid.New(),
		Quantity:  5,
	}

	_, err := db.NewInsert().
		Model(itemToRemove).
		Exec(ctx)

	require.NoError(t, err)

	_, err = db.NewInsert().
		Model(itemToKeep).
		Exec(ctx)

	require.NoError(t, err)

	err = repo.RemoveFromCart(ctx, userID, productID)

	require.NoError(t, err)

	var deleted domain.CartItem

	err = db.NewSelect().
		Model(&deleted).
		Where("id = ?", itemToRemove.ID).
		Scan(ctx)

	require.ErrorIs(t, err, sql.ErrNoRows)

	var remaining domain.CartItem

	err = db.NewSelect().
		Model(&remaining).
		Where("id = ?", itemToKeep.ID).
		Scan(ctx)

	require.NoError(t, err)
	require.Equal(t, itemToKeep.ID, remaining.ID)
	require.Equal(t, itemToKeep.ProductID, remaining.ProductID)
	require.Equal(t, itemToKeep.Quantity, remaining.Quantity)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.CartItem)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
	})
}

func TestOrderRepository_RemoveFromCart_NotFound(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	err := repo.RemoveFromCart(
		ctx,
		uuid.New(),
		uuid.New(),
	)

	require.ErrorIs(t, err, domain.ErrCartItemNotFound)
}

func TestOrderRepository_ClearCart(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()
	otherUserID := uuid.New()

	userItems := []domain.CartItem{
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: uuid.New(),
			Quantity:  2,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: uuid.New(),
			Quantity:  5,
		},
	}

	otherUserItem := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    otherUserID,
		ProductID: uuid.New(),
		Quantity:  1,
	}

	_, err := db.NewInsert().
		Model(&userItems).
		Exec(ctx)

	require.NoError(t, err)

	_, err = db.NewInsert().
		Model(otherUserItem).
		Exec(ctx)

	require.NoError(t, err)

	err = repo.ClearCart(ctx, userID)

	require.NoError(t, err)

	var remainingUserItems []domain.CartItem

	err = db.NewSelect().
		Model(&remainingUserItems).
		Where("user_id = ?", userID).
		Scan(ctx)

	require.NoError(t, err)
	require.Empty(t, remainingUserItems)

	var remainingOtherUserItem domain.CartItem

	err = db.NewSelect().
		Model(&remainingOtherUserItem).
		Where("id = ?", otherUserItem.ID).
		Scan(ctx)

	require.NoError(t, err)
	require.Equal(t, otherUserID, remainingOtherUserItem.UserID)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.CartItem)(nil)).
			Where("user_id IN (?, ?)", userID, otherUserID).
			Exec(ctx)
	})
}

func TestOrderRepository_CreateOrder(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		Total:        1500,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}

	err := repo.CreateOrder(ctx, order)

	require.NoError(t, err)

	var actual domain.Order

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", order.ID).
		Scan(ctx)

	require.NoError(t, err)

	require.Equal(t, order.ID, actual.ID)
	require.Equal(t, order.UserID, actual.UserID)
	require.Equal(t, order.Status, actual.Status)
	require.Equal(t, order.Total, actual.Total)
	require.WithinDuration(t, order.CreatedAt, actual.CreatedAt, time.Second)
	require.WithinDuration(t, order.DeliveryDate, actual.DeliveryDate, time.Second)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.Order)(nil)).
			Where("id = ?", order.ID).
			Exec(ctx)
	})
}

func TestOrderRepository_CreateOrderItems(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	orderID := uuid.New()

	order := &domain.Order{
		ID:           orderID,
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		Total:        1450, 
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}
	_, err := db.NewInsert().Model(order).Exec(ctx)
	require.NoError(t, err)

	items := []domain.OrderItem{
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 100,
			Quantity:     2,
		},
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 500,
			Quantity:     1,
		},
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 250,
			Quantity:     3,
		},
	}

	err = repo.CreateOrderItems(ctx, items)

	require.NoError(t, err)

	var actual []domain.OrderItem

	err = db.NewSelect().
		Model(&actual).
		Where("order_id = ?", orderID).
		Scan(ctx)

	require.NoError(t, err)
	require.Len(t, actual, 3)

	require.Equal(t, items[0].ID, actual[0].ID)
	require.Equal(t, items[0].OrderID, actual[0].OrderID)
	require.Equal(t, items[0].ProductID, actual[0].ProductID)
	require.Equal(t, items[0].ProductPrice, actual[0].ProductPrice)
	require.Equal(t, items[0].Quantity, actual[0].Quantity)

	require.Equal(t, items[1].ID, actual[1].ID)
	require.Equal(t, items[1].ProductID, actual[1].ProductID)
	require.Equal(t, items[1].ProductPrice, actual[1].ProductPrice)
	require.Equal(t, items[1].Quantity, actual[1].Quantity)

	require.Equal(t, items[2].ID, actual[2].ID)
	require.Equal(t, items[2].ProductID, actual[2].ProductID)
	require.Equal(t, items[2].ProductPrice, actual[2].ProductPrice)
	require.Equal(t, items[2].Quantity, actual[2].Quantity)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.OrderItem)(nil)).
			Where("order_id = ?", orderID).
			Exec(ctx)
	})
}

func TestOrderRepository_CreateOrderItems_Empty(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	err := repo.CreateOrderItems(
		context.Background(),
		[]domain.OrderItem{},
	)

	require.NoError(t, err)
}

func TestOrderRepository_GetOrderByID(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		Total:        1500,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
		DeliveryDate: time.Now().Add(48 * time.Hour).Truncate(time.Microsecond),
	}

	_, err := db.NewInsert().
		Model(order).
		Exec(ctx)

	require.NoError(t, err)

	result, err := repo.GetOrderByID(ctx, order.ID)

	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, order.ID, result.ID)
	require.Equal(t, order.UserID, result.UserID)
	require.Equal(t, order.Status, result.Status)
	require.Equal(t, order.Total, result.Total)
	require.WithinDuration(t, order.CreatedAt, result.CreatedAt, time.Second)
	require.WithinDuration(t, order.DeliveryDate, result.DeliveryDate, time.Second)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.Order)(nil)).
			Where("id = ?", order.ID).
			Exec(ctx)
	})
}

func TestOrderRepository_GetOrderByID_NotFound(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	order, err := repo.GetOrderByID(ctx, uuid.New())

	require.Nil(t, order)
	require.ErrorIs(t, err, domain.ErrOrderNotFound)
}

func TestOrderRepository_GetOrderItems(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	orderID := uuid.New()
	otherOrderID := uuid.New()

	order1 := &domain.Order{
		ID:           orderID,
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		Total:        700,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}
	_, err := db.NewInsert().Model(order1).Exec(ctx)
	require.NoError(t, err)

	order2 := &domain.Order{
		ID:           otherOrderID,
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		Total:        9990,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}
	_, err = db.NewInsert().Model(order2).Exec(ctx)
	require.NoError(t, err)

	items := []domain.OrderItem{
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 100,
			Quantity:     2,
		},
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 500,
			Quantity:     1,
		},
		{
			ID:           uuid.New(),
			OrderID:      otherOrderID,
			ProductID:    uuid.New(),
			ProductPrice: 999,
			Quantity:     10,
		},
	}

	_, err = db.NewInsert().
		Model(&items).
		Exec(ctx)

	require.NoError(t, err)

	result, err := repo.GetOrderItems(ctx, orderID)

	require.NoError(t, err)
	require.Len(t, result, 2)

	require.Equal(t, items[0].ID, result[0].ID)
	require.Equal(t, items[0].OrderID, result[0].OrderID)
	require.Equal(t, items[0].ProductID, result[0].ProductID)
	require.Equal(t, items[0].ProductPrice, result[0].ProductPrice)
	require.Equal(t, items[0].Quantity, result[0].Quantity)

	require.Equal(t, items[1].ID, result[1].ID)
	require.Equal(t, items[1].OrderID, result[1].OrderID)
	require.Equal(t, items[1].ProductID, result[1].ProductID)
	require.Equal(t, items[1].ProductPrice, result[1].ProductPrice)
	require.Equal(t, items[1].Quantity, result[1].Quantity)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.OrderItem)(nil)).
			Where("order_id IN (?, ?)", orderID, otherOrderID).
			Exec(ctx)
	})
}

func TestOrderRepository_GetOrderItems_Empty(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	items, err := repo.GetOrderItems(ctx, uuid.New())

	require.NoError(t, err)
	require.Empty(t, items)
}

func TestOrderRepository_Transaction_Commit(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		Total:        2000,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}

	err := repo.Transaction(ctx, func(txRepo domain.OrderRepository) error {
		return txRepo.CreateOrder(ctx, order)
	})

	require.NoError(t, err)

	var actual domain.Order

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", order.ID).
		Scan(ctx)

	require.NoError(t, err)

	require.Equal(t, order.ID, actual.ID)
	require.Equal(t, order.UserID, actual.UserID)
	require.Equal(t, order.Status, actual.Status)
	require.Equal(t, order.Total, actual.Total)

	t.Cleanup(func() {
		_, _ = db.NewDelete().
			Model((*domain.Order)(nil)).
			Where("id = ?", order.ID).
			Exec(ctx)
	})
}

func TestOrderRepository_Transaction_Rollback(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		Total:        3000,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}

	expectedErr := errors.New("something went wrong")

	err := repo.Transaction(ctx, func(txRepo domain.OrderRepository) error {
		err := txRepo.CreateOrder(ctx, order)
		require.NoError(t, err)

		return expectedErr
	})

	require.ErrorIs(t, err, expectedErr)

	var actual domain.Order

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", order.ID).
		Scan(ctx)

	require.ErrorIs(t, err, sql.ErrNoRows)
}