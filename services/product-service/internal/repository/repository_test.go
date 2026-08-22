package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	testtools "github.com/byorty/test-marketplace/services/common/test-tools"
	"github.com/byorty/test-marketplace/services/product-service/internal/domain"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	_ "github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
)


func newProductTestRepository(t *testing.T) (*ProductRepository, *bun.DB) {
	t.Helper()

	database := testtools.NewTestDB(t)

	repo := New(
		database,
		zap.NewNop(),
	)

	return repo, database
}

func newTestProduct() *domain.Product {
	now := time.Now()

	return &domain.Product{
		ID:           uuid.New(),
		Name:         "iPhone",
		Description:  "Phone",
		Category:     "Electronics",
		Price:        100,
		DeliveryDays: 2,
		Rating:       4.5,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestRepository_Create(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	product := newTestProduct()

	err := repo.Create(ctx, product)

	require.NoError(t, err)

	var actual domain.Product

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", product.ID).
		Scan(ctx)

	require.NoError(t, err)

	require.Equal(t, product.ID, actual.ID)
	require.Equal(t, product.Name, actual.Name)
	require.Equal(t, product.Description, actual.Description)
	require.Equal(t, product.Category, actual.Category)
	require.Equal(t, product.Price, actual.Price)
	require.Equal(t, product.DeliveryDays, actual.DeliveryDays)
	require.Equal(t, product.Rating, actual.Rating)
}

func TestRepository_GetByID(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	product := newTestProduct()

	_, err := db.NewInsert().
		Model(product).
		Exec(ctx)

	require.NoError(t, err)

	actual, err := repo.GetByID(ctx, product.ID)

	require.NoError(t, err)
	require.NotNil(t, actual)

	require.Equal(t, product.ID, actual.ID)
	require.Equal(t, product.Name, actual.Name)
	require.Equal(t, product.Description, actual.Description)
	require.Equal(t, product.Category, actual.Category)
	require.Equal(t, product.Price, actual.Price)
	require.Equal(t, product.DeliveryDays, actual.DeliveryDays)
	require.Equal(t, product.Rating, actual.Rating)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	repo, _ := newProductTestRepository(t)

	product, err := repo.GetByID(
		context.Background(),
		uuid.New(),
	)

	require.Nil(t, product)
	require.ErrorIs(t, err, domain.ErrProductNotFound)
}

func TestRepository_Update(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	product := newTestProduct()

	product.CreatedAt = time.Now().Add(-time.Hour)
	product.UpdatedAt = product.CreatedAt

	_, err := db.NewInsert().
		Model(product).
		Exec(ctx)

	require.NoError(t, err)

	product.Name = "iPhone 17"
	product.Description = "New description"
	product.Price = 150
	product.DeliveryDays = 3
	product.Rating = 5

	oldUpdatedAt := product.UpdatedAt

	updated, err := repo.Update(ctx, product)

	require.NoError(t, err)
	require.NotNil(t, updated)

	require.Equal(t, product.ID, updated.ID)
	require.Equal(t, "iPhone 17", updated.Name)
	require.Equal(t, "New description", updated.Description)
	require.Equal(t, "Electronics", updated.Category)
	require.Equal(t, int64(150), updated.Price)
	require.Equal(t, 3, updated.DeliveryDays)
	require.Equal(t, float32(5), updated.Rating)

	require.True(t, updated.UpdatedAt.After(oldUpdatedAt))
}

func TestRepository_Update_NotFound(t *testing.T) {
	repo, _ := newProductTestRepository(t)

	p := &domain.Product{
		ID:           uuid.New(),
		Name:         "iPhone",
		Description:  "Phone",
		Category:     "Electronics",
		Price:        100,
		DeliveryDays: 2,
		Rating:       5,
	}

	updated, err := repo.Update(context.Background(), p)

	require.Nil(t, updated)
	require.ErrorIs(t, err, domain.ErrProductNotFound)
}

func TestRepository_Delete(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	product := newTestProduct()

	_, err := db.NewInsert().
		Model(product).
		Exec(ctx)

	require.NoError(t, err)

	err = repo.Delete(ctx, product.ID)

	require.NoError(t, err)

	var actual domain.Product

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", product.ID).
		Scan(ctx)

	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestRepository_Delete_NotFound(t *testing.T) {
	repo, _ := newProductTestRepository(t)

	err := repo.Delete(
		context.Background(),
		uuid.New(),
	)

	require.ErrorIs(t, err, domain.ErrProductNotFound)
}

func TestRepository_List(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	products := []*domain.Product{
		{
			ID:           uuid.New(),
			Name:         "iPhone 17",
			Description:  "Apple phone",
			Category:     "Electronics",
			Price:        150,
			DeliveryDays: 2,
			Rating:       5,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Name:         "MacBook",
			Description:  "Apple laptop",
			Category:     "Electronics",
			Price:        2000,
			DeliveryDays: 5,
			Rating:       4.8,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Name:         "T-Shirt",
			Description:  "Cotton t-shirt",
			Category:     "Clothes",
			Price:        30,
			DeliveryDays: 3,
			Rating:       4.2,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, product := range products {
		_, err := db.NewInsert().
			Model(product).
			Exec(ctx)

		require.NoError(t, err)
	}

	result, err := repo.List(
		ctx,
		domain.ListFilter{},
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, int64(3), result.Total)
	require.Equal(t, 1, result.Page)
	require.Equal(t, 20, result.PageSize)
	require.Len(t, result.Items, 3)
}

func TestRepository_List_FilterAndSortByPrice(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	products := []*domain.Product{
		{
			ID:           uuid.New(),
			Name:         "Cheap",
			Description:  "Cheap product",
			Category:     "Electronics",
			Price:        100,
			DeliveryDays: 2,
			Rating:       4,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Name:         "Medium",
			Description:  "Medium product",
			Category:     "Electronics",
			Price:        500,
			DeliveryDays: 3,
			Rating:       4.5,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Name:         "Expensive",
			Description:  "Expensive product",
			Category:     "Electronics",
			Price:        1000,
			DeliveryDays: 5,
			Rating:       5,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, product := range products {
		_, err := db.NewInsert().
			Model(product).
			Exec(ctx)

		require.NoError(t, err)
	}

	minPrice := int64(200)

	result, err := repo.List(
		ctx,
		domain.ListFilter{
			MinPrice: &minPrice,
			SortBy:   domain.SortByPrice,
			Order:    domain.Asc,
		},
	)

	require.NoError(t, err)

	require.Equal(t, int64(2), result.Total)
	require.Len(t, result.Items, 2)

	require.Equal(t, int64(500), result.Items[0].Price)
	require.Equal(t, int64(1000), result.Items[1].Price)
}

func TestRepository_List_Pagination(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		product := &domain.Product{
			ID:           uuid.New(),
			Name:         "Product " + string(rune('0'+i)),
			Description:  "Product",
			Category:     "Electronics",
			Price:        int64(i * 100),
			DeliveryDays: 2,
			Rating:       5,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		_, err := db.NewInsert().
			Model(product).
			Exec(ctx)

		require.NoError(t, err)
	}

	result, err := repo.List(
		ctx,
		domain.ListFilter{
			Page:     2,
			PageSize: 2,
			SortBy:   domain.SortByPrice,
			Order:    domain.Asc,
		},
	)

	require.NoError(t, err)

	require.Equal(t, int64(5), result.Total)
	require.Equal(t, 2, result.Page)
	require.Equal(t, 2, result.PageSize)
	require.Len(t, result.Items, 2)

	require.Equal(t, int64(200), result.Items[0].Price)
	require.Equal(t, int64(300), result.Items[1].Price)
}