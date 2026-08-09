package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	return db
}

func newProductTestRepository(t *testing.T) (*ProductRepository, *bun.DB) {
	t.Helper()

	db := newTestDB(t)

	repo := New(
		db,
		zap.NewNop(),
	)

	return repo, db
}

func TestRepository_Create(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	p := &domain.Product{
		ID:           uuid.New(),
		Name:         "iPhone",
		Description:  "Phone",
		Category:     "Electronics",
		Price:        100,
		DeliveryDays: 2,
		Rating:       0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, p)

	require.NoError(t, err)

	var actual domain.Product

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", p.ID).
		Scan(ctx)

	require.NoError(t, err)
	require.Equal(t, p.ID, actual.ID)
	require.Equal(t, p.Name, actual.Name)
	require.Equal(t, p.Description, actual.Description)
	require.Equal(t, p.Category, actual.Category)
	require.Equal(t, p.Price, actual.Price)
	require.Equal(t, p.DeliveryDays, actual.DeliveryDays)
	require.Equal(t, p.Rating, actual.Rating)
}

func TestRepository_GetByID(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	p := &domain.Product{
		ID:           uuid.New(),
		Name:         "iPhone",
		Description:  "Phone",
		Category:     "Electronics",
		Price:        100,
		DeliveryDays: 3,
		Rating:       5,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := db.NewInsert().
		Model(p).
		Exec(ctx)

	require.NoError(t, err)

	product, err := repo.GetByID(ctx, p.ID)

	require.NoError(t, err)
	require.Equal(t, p.ID, product.ID)
	require.Equal(t, p.Name, product.Name)
	require.Equal(t, p.Description, product.Description)
	require.Equal(t, p.Category, product.Category)
	require.Equal(t, p.Price, product.Price)
	require.Equal(t, p.DeliveryDays, product.DeliveryDays)
	require.Equal(t, p.Rating, product.Rating)
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

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := createdAt

	p := &domain.Product{
		ID:           uuid.New(),
		Name:         "iPhone",
		Description:  "Old description",
		Category:     "Electronics",
		Price:        100,
		DeliveryDays: 3,
		Rating:       4.5,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	_, err := db.NewInsert().
		Model(p).
		Exec(ctx)

	require.NoError(t, err)

	p.Name = "iPhone 17"
	p.Description = "New description"
	p.Price = 150
	p.DeliveryDays = 2
	p.Rating = 5.0

	updated, err := repo.Update(ctx, p)

	require.NoError(t, err)
	require.NotNil(t, updated)

	require.Equal(t, p.ID, updated.ID)
	require.Equal(t, "iPhone 17", updated.Name)
	require.Equal(t, "New description", updated.Description)
	require.Equal(t, "Electronics", updated.Category)
	require.Equal(t, int64(150), updated.Price)
	require.Equal(t, 2, updated.DeliveryDays)
	require.Equal(t, 5.0, updated.Rating)

	require.True(t, updated.UpdatedAt.After(updatedAt))
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

	p := &domain.Product{
		ID:           uuid.New(),
		Name:         "iPhone",
		Description:  "Phone",
		Category:     "Electronics",
		Price:        100,
		DeliveryDays: 2,
		Rating:       5,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := db.NewInsert().
		Model(p).
		Exec(ctx)

	require.NoError(t, err)

	err = repo.Delete(ctx, p.ID)

	require.NoError(t, err)

	var actual domain.Product

	err = db.NewSelect().
		Model(&actual).
		Where("id = ?", p.ID).
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

	_, err := db.NewDelete().
		Model((*domain.Product)(nil)).
		Where("1=1").
		Exec(ctx)
	require.NoError(t, err)

	products := []*domain.Product{
		{
			ID:           uuid.New(),
			Name:         "iPhone 17",
			Description:  "Apple phone",
			Category:     "Electronics",
			Price:        150,
			DeliveryDays: 2,
			Rating:       5.0,
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

	for _, p := range products {
		_, err := db.NewInsert().
			Model(p).
			Exec(ctx)

		require.NoError(t, err)
	}

	result, err := repo.List(ctx, domain.ListFilter{})

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

	_, err := db.NewDelete().
		Model((*domain.Product)(nil)).
		Where("1=1").
		Exec(ctx)
	require.NoError(t, err)

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

	for _, p := range products {
		_, err := db.NewInsert().
			Model(p).
			Exec(ctx)

		require.NoError(t, err)
	}

	minPrice := int64(200)

	result, err := repo.List(ctx, domain.ListFilter{
		MinPrice: &minPrice,
		SortBy:   domain.SortByPrice,
		Order:    domain.Asc,
	})

	require.NoError(t, err)

	require.Equal(t, int64(2), result.Total)
	require.Len(t, result.Items, 2)

	require.Equal(t, int64(500), result.Items[0].Price)
	require.Equal(t, int64(1000), result.Items[1].Price)
}

func TestRepository_List_Pagination(t *testing.T) {
	repo, db := newProductTestRepository(t)

	ctx := context.Background()

	_, err := db.NewDelete().
		Model((*domain.Product)(nil)).
		Where("1=1").
		Exec(ctx)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		p := &domain.Product{
			ID:           uuid.New(),
			Name:         fmt.Sprintf("Product %d", i),
			Description:  "Product",
			Category:     "Electronics",
			Price:        int64(i * 100),
			DeliveryDays: 2,
			Rating:       5,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		_, err := db.NewInsert().
			Model(p).
			Exec(ctx)

		require.NoError(t, err)
	}

	result, err := repo.List(ctx, domain.ListFilter{
		Page:     2,
		PageSize: 2,
		SortBy:   domain.SortByPrice,
		Order:    domain.Asc,
	})

	require.NoError(t, err)

	require.Equal(t, int64(5), result.Total)
	require.Equal(t, 2, result.Page)
	require.Equal(t, 2, result.PageSize)
	require.Len(t, result.Items, 2)

	require.Equal(t, int64(200), result.Items[0].Price)
	require.Equal(t, int64(300), result.Items[1].Price)
}