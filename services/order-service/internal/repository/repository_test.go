package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	testtools "github.com/byorty/test-marketplace/services/common/test-tools"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	_ "github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
)


func newOrderTestRepository(t *testing.T) (*OrderRepository, *bun.DB) {
	t.Helper()

	database := testtools.NewTestDB(t)

	repo := New(
		database,
		zap.NewNop(),
	)

	return repo, database
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

	err := repo.AddToCart(ctx, userID, item)

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

	for i := range items {
		_, err := db.NewInsert().
			Model(&items[i]).
			Exec(ctx)

		require.NoError(t, err)
	}

	actual, err := repo.GetCart(ctx, userID)

	require.NoError(t, err)
	require.Len(t, actual, 2)

	require.ElementsMatch(t, items, actual)
}

func TestOrderRepository_GetCart_Empty(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()

	items, err := repo.GetCart(ctx, userID)

	require.ErrorIs(t, err, domain.ErrCartEmpty)
	require.Nil(t, items)
}

func TestOrderRepository_GetCartItem(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	item := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ProductID: uuid.New(),
		Quantity:  3,
	}

	_, err := db.NewInsert().
		Model(item).
		Exec(ctx)

	require.NoError(t, err)

	actual, err := repo.GetCartItem(
		ctx,
		item.UserID,
		item.ProductID,
	)

	require.NoError(t, err)
	require.Equal(t, item, actual)
}

func TestOrderRepository_GetCartItem_NotFound(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	item, err := repo.GetCartItem(
		ctx,
		uuid.New(),
		uuid.New(),
	)

	require.ErrorIs(t, err, domain.ErrCartItemNotFound)
	require.Nil(t, item)
}

func TestOrderRepository_RemoveFromCart(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()
	item := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: uuid.New(),
		Quantity:  2,
	}

	err := repo.AddToCart(ctx, userID, item)
	require.NoError(t, err)

	err = repo.RemoveFromCart(ctx, userID, item.ID)
	require.NoError(t, err)

	var actual domain.CartItem

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", item.ID).
		Scan(ctx)

	require.ErrorIs(t, err, sql.ErrNoRows)
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

func TestOrderRepository_RemoveFromCart_WrongUser(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()

	item := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: uuid.New(),
		Quantity:  1,
	}

	err := repo.AddToCart(ctx, userID, item)
	require.NoError(t, err)

	err = repo.RemoveFromCart(
		ctx,
		uuid.New(),
		item.ID,
	)

	require.ErrorIs(t, err, domain.ErrCartItemNotFound)
}

func TestOrderRepository_ClearCart(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	userID := uuid.New()

	items := []domain.CartItem{
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: uuid.New(),
			Quantity:  1,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: uuid.New(),
			Quantity:  3,
		},
	}

	for i := range items {
		err := repo.AddToCart(ctx, userID, &items[i])
		require.NoError(t, err)
	}

	otherItem := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ProductID: uuid.New(),
		Quantity:  1,
	}

	err := repo.AddToCart(ctx, otherItem.UserID, otherItem)
	require.NoError(t, err)

	err = repo.ClearCart(ctx, userID)
	require.NoError(t, err)

	var actual []domain.CartItem

	err = db.NewSelect().
		Model(&actual).
		Where("user_id = ?", userID).
		Scan(ctx)

	require.NoError(t, err)
	require.Empty(t, actual)

	var other domain.CartItem

	err = db.NewSelect().
		Model(&other).
		Where("id = ?", otherItem.ID).
		Scan(ctx)

	require.NoError(t, err)
	require.Equal(t, otherItem.ID, other.ID)
}

func TestOrderRepository_CreateOrder(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		TotalPrice:   1500,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
		Items: []domain.OrderItem{
			{
				ID:           uuid.New(),
				ProductID:    uuid.New(),
				ProductPrice: 1000,
				Quantity:     1,
			},
			{
				ID:           uuid.New(),
				ProductID:    uuid.New(),
				ProductPrice: 500,
				Quantity:     1,
			},
		},
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
	require.Equal(t, order.TotalPrice, actual.TotalPrice)
	require.WithinDuration(
		t,
		order.CreatedAt,
		actual.CreatedAt,
		time.Second,
	)
	require.WithinDuration(
		t,
		order.DeliveryDate,
		actual.DeliveryDate,
		time.Second,
	)

	var actualItems []domain.OrderItem

	err = db.NewSelect().
		Model(&actualItems).
		Where("order_id = ?", order.ID).
		Order("id").
		Scan(ctx)

	require.NoError(t, err)
	require.Len(t, actualItems, 2)

	require.Equal(t, order.Items[0].OrderID, actualItems[0].OrderID)
	require.Equal(t, order.Items[0].ProductID, actualItems[0].ProductID)
	require.Equal(t, order.Items[0].ProductPrice, actualItems[0].ProductPrice)
	require.Equal(t, order.Items[0].Quantity, actualItems[0].Quantity)

	require.Equal(t, order.Items[1].ID, actualItems[1].ID)
	require.Equal(t, order.Items[1].OrderID, actualItems[1].OrderID)
	require.Equal(t, order.Items[1].ProductID, actualItems[1].ProductID)
	require.Equal(t, order.Items[1].ProductPrice, actualItems[1].ProductPrice)
	require.Equal(t, order.Items[1].Quantity, actualItems[1].Quantity)
}

func TestOrderRepository_CreateOrderItems(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	orderID := uuid.New()

	order := &domain.Order{
		ID:           orderID,
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		TotalPrice:   2000,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}

	err := repo.CreateOrder(ctx, order)
	require.NoError(t, err)

	items := []domain.OrderItem{
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 1200,
			Quantity:     1,
		},
		{
			ID:           uuid.New(),
			OrderID:      orderID,
			ProductID:    uuid.New(),
			ProductPrice: 800,
			Quantity:     1,
		},
	}

	err = repo.CreateOrderItems(ctx, items)
	require.NoError(t, err)

	var actual []domain.OrderItem

	err = db.NewSelect().
		Model(&actual).
		Where("order_id = ?", orderID).
		Order("id").
		Scan(ctx)

	require.NoError(t, err)
	require.Len(t, actual, 2)

	require.Equal(t, items[0].OrderID, actual[0].OrderID)
	require.Equal(t, items[0].ProductID, actual[0].ProductID)
	require.Equal(t, items[0].ProductPrice, actual[0].ProductPrice)
	require.Equal(t, items[0].Quantity, actual[0].Quantity)

	require.Equal(t, items[1].ID, actual[1].ID)
	require.Equal(t, items[1].OrderID, actual[1].OrderID)
	require.Equal(t, items[1].ProductID, actual[1].ProductID)
	require.Equal(t, items[1].ProductPrice, actual[1].ProductPrice)
	require.Equal(t, items[1].Quantity, actual[1].Quantity)
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
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		TotalPrice:   1500,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
		DeliveryDate: time.Now().Add(48 * time.Hour).Truncate(time.Microsecond),
	}

	err := repo.CreateOrder(ctx, order)
	require.NoError(t, err)

	actual, err := repo.GetOrderByID(ctx, order.ID)

	require.NoError(t, err)
	require.NotNil(t, actual)

	require.Equal(t, order.ID, actual.ID)
	require.Equal(t, order.UserID, actual.UserID)
	require.Equal(t, order.Status, actual.Status)
	require.Equal(t, order.TotalPrice, actual.TotalPrice)
	require.WithinDuration(t, order.CreatedAt, actual.CreatedAt, time.Second)
	require.WithinDuration(t, order.DeliveryDate, actual.DeliveryDate, time.Second)

	require.Empty(t, actual.Items)
}

func TestOrderRepository_GetOrderByID_NotFound(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	actual, err := repo.GetOrderByID(ctx, uuid.New())

	require.Nil(t, actual)
	require.ErrorIs(t, err, domain.ErrOrderNotFound)
}

func TestOrderRepository_GetOrderItems(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		TotalPrice:   2000,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}

	err := repo.CreateOrder(ctx, order)
	require.NoError(t, err)

	items := []domain.OrderItem{
		{
			ID:           uuid.New(),
			OrderID:      order.ID,
			ProductID:    uuid.New(),
			ProductPrice: 1200,
			Quantity:     1,
		},
		{
			ID:           uuid.New(),
			OrderID:      order.ID,
			ProductID:    uuid.New(),
			ProductPrice: 800,
			Quantity:     1,
		},
	}

	err = repo.CreateOrderItems(ctx, items)
	require.NoError(t, err)

	actual, err := repo.GetOrderItems(ctx, order.ID)

	require.NoError(t, err)
	require.Len(t, actual, 2)

	expectedByID := make(map[uuid.UUID]domain.OrderItem, len(items))

	for _, item := range items {
		expectedByID[item.ID] = item
	}

	actualByID := make(map[uuid.UUID]domain.OrderItem, len(actual))

	for _, item := range actual {
		actualByID[item.ID] = item
	}

	require.Equal(t, expectedByID, actualByID)
}

func TestOrderRepository_GetOrderItems_Empty(t *testing.T) {
	repo, _ := newOrderTestRepository(t)

	ctx := context.Background()

	actual, err := repo.GetOrderItems(ctx, uuid.New())

	require.NoError(t, err)
	require.Empty(t, actual)
}

func TestOrderRepository_Transaction_Commit(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		TotalPrice:   1500,
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
	require.Equal(t, order.TotalPrice, actual.TotalPrice)
}

func TestOrderRepository_Transaction_Rollback(t *testing.T) {
	repo, db := newOrderTestRepository(t)

	ctx := context.Background()

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Status:       domain.StatusCreated,
		TotalPrice:   1500,
		CreatedAt:    time.Now(),
		DeliveryDate: time.Now().Add(48 * time.Hour),
	}

	expectedErr := errors.New("rollback transaction")

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