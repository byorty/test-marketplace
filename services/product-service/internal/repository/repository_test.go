package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newProductTestRepository(t *testing.T) (*ProductRepository, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}

	repo := &ProductRepository{
		db:  mock,
		log: zap.NewNop(),
	}

	return repo, mock
}

func TestRepository_Create(t *testing.T) {
    t.Parallel()

    repo, mock := newProductTestRepository(t)

    p := &domain.Product{
        ID: uuid.New(),
        Name: "iPhone",
        Description: "Phone",
        Category: "Electronics",
        Price: 100,
        DeliveryDays: 2,
        Rating: 0,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    mock.ExpectExec("INSERT INTO products").
        WithArgs(
            p.ID,
            p.Name,
            p.Description,
            p.Category,
            p.Price,
            p.DeliveryDays,
            p.Rating,
            p.CreatedAt,
            p.UpdatedAt,
        ).
        WillReturnResult(
            pgxmock.NewResult("INSERT", 1),
        )

    err := repo.Create(context.Background(), p)

    require.NoError(t, err)
    require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID(t *testing.T) {

	repo, mock := newProductTestRepository(t)

	id := uuid.New()

	rows := pgxmock.NewRows([]string{
		"id",
		"name",
		"description",
		"category",
		"price",
		"delivery_days",
		"rating",
		"created_at",
		"updated_at",
	})

	now := time.Now()

	rows.AddRow(
		id,
		"iPhone",
		"Phone",
		"Electronics",
		int64(100),
		3,
		5.0,
		now,
		now,
	)

	mock.ExpectQuery("SELECT").
		WithArgs(id).
		WillReturnRows(rows)

	product, err := repo.GetByID(context.Background(), id)

	require.NoError(t, err)
	require.Equal(t, id, product.ID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
    repo, mock := newProductTestRepository(t)

    product := &domain.Product{
        ID:           uuid.New(),
        Name:         "Mac",
        Description:  "Laptop",
        Category:     "Electronics",
        Price:        500,
        DeliveryDays: 5,
        Rating:       4.9,
    }

    rows := pgxmock.NewRows([]string{
        "id",
        "name",
        "description",
        "category",
        "price",
        "delivery_days",
        "rating",
        "created_at",
        "updated_at",
    })

    now := time.Now()

    rows.AddRow(
        product.ID,
        product.Name,
        product.Description,
        product.Category,
        product.Price,
        product.DeliveryDays,
        product.Rating,
        now,
        now,
    )

    mock.ExpectQuery("UPDATE products").
        WithArgs(
            product.Name,
            product.Description,
            product.Category,
            product.Price,
            product.DeliveryDays,
            product.Rating,
            pgxmock.AnyArg(),
            product.ID,
        ).
        WillReturnRows(rows)

    updated, err := repo.Update(context.Background(), product)

    require.NoError(t, err)
    require.Equal(t, product.Name, updated.Name)
    require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete(t *testing.T) {

	repo, mock := newProductTestRepository(t)

	id := uuid.New()

	mock.ExpectExec("DELETE FROM products").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(context.Background(), id)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_List(t *testing.T) {
	t.Parallel()

	repo, mock := newProductTestRepository(t)

	now := time.Now()

	minPrice := int64(1000)

	filter := domain.ListFilter{
		Name:     "iphone",
		MinPrice: &minPrice,
		Page:     1,
		PageSize: 20,
		SortBy:   domain.SortByPrice,
		Order:    domain.Asc,
	}

	tests := []struct {
		name    string
		prepare func()
		check   func(t *testing.T, list *domain.ProductList)
		wantErr error
	}{
		{
			name: "success",
			prepare: func() {
				countRows := pgxmock.NewRows([]string{"count"}).
					AddRow(int64(2))

				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs("%iphone%", int64(1000)).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id",
					"name",
					"description",
					"category",
					"price",
					"delivery_days",
					"rating",
					"created_at",
					"updated_at",
				})

				rows.AddRow(
					uuid.New(),
					"iPhone 16",
					"Apple",
					"Phones",
					int64(120000),
					3,
					4.9,
					now,
					now,
				)

				rows.AddRow(
					uuid.New(),
					"iPhone 15",
					"Apple",
					"Phones",
					int64(90000),
					2,
					4.7,
					now,
					now,
				)

				mock.ExpectQuery(`SELECT`).
					WithArgs("%iphone%", int64(1000)).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, list *domain.ProductList) {
				require.NotNil(t, list)

				require.Equal(t, int64(2), list.Total)
				require.Equal(t, 1, list.Page)
				require.Equal(t, 20, list.PageSize)

				require.Len(t, list.Items, 2)

				require.Equal(t, "iPhone 16", list.Items[0].Name)
				require.Equal(t, int64(120000), list.Items[0].Price)

				require.Equal(t, "iPhone 15", list.Items[1].Name)
				require.Equal(t, int64(90000), list.Items[1].Price)
			},
		},
		{
			name: "count query error",
			prepare: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs("%iphone%", int64(1000)).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "select query error",
			prepare: func() {
				countRows := pgxmock.NewRows([]string{"count"}).
					AddRow(int64(1))

				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs("%iphone%", int64(1000)).
					WillReturnRows(countRows)

				mock.ExpectQuery(`SELECT`).
					WithArgs("%iphone%", int64(1000)).
					WillReturnError(errors.New("query error"))
			},
			wantErr: errors.New("query error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()

			list, err := repo.List(context.Background(), filter)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr.Error())
				require.Nil(t, list)
			} else {
				require.NoError(t, err)
				tt.check(t, list)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}