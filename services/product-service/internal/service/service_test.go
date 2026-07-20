package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
    return &s
}

func int64Ptr(i int64) *int64 {
    return &i
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestService(repo domain.Repository) *Service {
	return &Service{
		repo: repo,
		log: newTestLogger(),
	}
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input *CreateProduct
		mock *MockRepository
		checkResult func(t *testing.T, p *domain.Product)
		wantErr error
	}{
		{
			name: "success",

			input: &CreateProduct{
				Name: "iPhone 17",
				Description: "Best phone",
				Category: "Electronics",
				Price: 120000,
				DeliveryDays: 3,
			},

			mock: &MockRepository{
				CreateFunc: func(ctx context.Context, p *domain.Product) error {
					return nil
				},
			},

			checkResult: func(t *testing.T, p *domain.Product) {
				require.NotNil(t, p)

				require.NotEqual(t, uuid.Nil, p.ID)

				require.Equal(t, "iPhone 17", p.Name)
				require.Equal(t, "Best phone", p.Description)
				require.Equal(t, "Electronics", p.Category)
				require.Equal(t, int64(120000), p.Price)
				require.Equal(t, 3, p.DeliveryDays)

				require.Equal(t, float64(0), p.Rating)

				require.False(t, p.CreatedAt.IsZero())
				require.False(t, p.UpdatedAt.IsZero())
			},
		},
		{
			name: "nil input",
			input: nil,
			mock: &MockRepository{},
			wantErr: ErrNilInput,
		},
		{
			name: "empty name",
			input: &CreateProduct{
				Name: "",
			},
			mock: &MockRepository{},
			wantErr: ErrInvalidProductName,
		},
		{
			name: "repository error",
			input: &CreateProduct{
				Name: "iPhone",
				Description: "Phone",
				Category: "Electronics",
				Price: 70000,
				DeliveryDays: 2,
			},
			mock: &MockRepository{
				CreateFunc: func(ctx context.Context, p *domain.Product) error {
					return domain.ErrProductNotFound
				},
			},
			wantErr: domain.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock)

			product, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, product)
				return
			}

			require.NoError(t, err)

			tt.checkResult(t, product)

			require.Equal(t, 1, tt.mock.CreateCalls)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repository error")

	product := &domain.Product{
	ID:            uuid.New(),
	Name:          "iPhone 17",
	Description:   "Best phone",
	Category:      "Electronics",
	Price:         120000,
	DeliveryDays:  3,
	Rating:        4.8,
	CreatedAt:     time.Now(),
	UpdatedAt:     time.Now(),
}

	tests := []struct {
		name string
		id uuid.UUID
		mock *MockRepository
		wantErr error
		checkResult func(t *testing.T, p *domain.Product)
	}{
		{
			name: "success",

			id: product.ID,

			mock: &MockRepository{
				GetByIDFunc: func(ctx context.Context, u uuid.UUID) (*domain.Product, error) {
					return product, nil
				},
			},

			checkResult: func(t *testing.T, p *domain.Product) {
				require.NotNil(t, p)

				require.Equal(t, product.ID, p.ID)
				require.Equal(t, product.Name, p.Name)
				require.Equal(t, product.Description, p.Description)
				require.Equal(t, product.Category, p.Category)
				require.Equal(t, product.Price, p.Price)
				require.Equal(t, product.DeliveryDays, p.DeliveryDays)
				require.Equal(t, product.Rating, p.Rating)
				require.Equal(t, product.CreatedAt, p.CreatedAt)
				require.Equal(t, product.UpdatedAt, p.UpdatedAt)
			},
		},
		{
			name: "invalid id",
			id: uuid.Nil,
			mock: &MockRepository{},
			wantErr: ErrInvalidID,
		},
		{
			name: "product not found",
			id: uuid.New(),
			mock: &MockRepository{
				GetByIDFunc: func(ctx context.Context, u uuid.UUID) (*domain.Product, error) {
					return nil, domain.ErrProductNotFound
				},
			},
			wantErr: domain.ErrProductNotFound,
		},
		{
			name: "repository error",
			id: uuid.New(),
			mock: &MockRepository{
				GetByIDFunc: func(ctx context.Context, u uuid.UUID) (*domain.Product, error) {
					return nil, repoErr
				},
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock)

			result, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)

				if tt.id == uuid.Nil {
					require.Zero(t, tt.mock.GetByIDCalls)
				} else {
					require.Equal(t, 1, tt.mock.GetByIDCalls)
				}

				require.Zero(t, tt.mock.CreateCalls)
				require.Zero(t, tt.mock.UpdateCalls)
				require.Zero(t, tt.mock.DeleteCalls)
				require.Zero(t, tt.mock.ListCalls)

				return
			}

			require.NoError(t, err)

			tt.checkResult(t, result)

			require.Equal(t, 1, tt.mock.GetByIDCalls)

			require.Zero(t, tt.mock.CreateCalls)
			require.Zero(t, tt.mock.UpdateCalls)
			require.Zero(t, tt.mock.DeleteCalls)
			require.Zero(t, tt.mock.ListCalls)
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repository error")

	tests := []struct {
		name string
		id uuid.UUID
		mock *MockRepository
		wantErr error
	}{
		{
			name: "success",
			id: uuid.New(),
			mock: &MockRepository{
				DeleteFunc: func(ctx context.Context, u uuid.UUID) error {
					return nil
				},
			},
		},
		{
			name: "invalid id",
			id: uuid.Nil,
			mock: &MockRepository{},
			wantErr: ErrInvalidID,
		},
		{
			name: "product not found",
			id: uuid.New(),
			mock: &MockRepository{
				DeleteFunc: func(ctx context.Context, u uuid.UUID) error {
					return domain.ErrProductNotFound
				},
			},
			wantErr: domain.ErrProductNotFound,
		},
		{
			name: "repository error",
			id: uuid.New(),
			mock: &MockRepository{
				DeleteFunc: func(ctx context.Context, u uuid.UUID) error {
					return repoErr
				},
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock)

			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				if tt.id == uuid.Nil {
					require.Zero(t, tt.mock.DeleteCalls)
				} else {
					require.Equal(t, 1, tt.mock.DeleteCalls)
				}

				require.Zero(t, tt.mock.CreateCalls)
				require.Zero(t, tt.mock.GetByIDCalls)
				require.Zero(t, tt.mock.UpdateCalls)
				require.Zero(t, tt.mock.ListCalls)

				return
			}

			require.NoError(t, err)

			require.Equal(t, 1, tt.mock.DeleteCalls)

			require.Zero(t, tt.mock.CreateCalls)
			require.Zero(t, tt.mock.GetByIDCalls)
			require.Zero(t, tt.mock.UpdateCalls)
			require.Zero(t, tt.mock.ListCalls)
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repository error")

	tests := []struct {
		name string
		filter domain.ListFilter
		mock *MockRepository
		wantErr error
		checkResult func(t *testing.T, result *domain.ProductList)
	}{
		{
			name: "success",

			filter: domain.ListFilter{
				Page: 2,
				PageSize: 10,
			},

			mock: &MockRepository{
				ListFunc: func(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
					require.Equal(t, 2, filter.Page)
					require.Equal(t, 10, filter.PageSize)

					return &domain.ProductList{
						Items: []*domain.Product{
							{
								ID: uuid.New(),
								Name: "iPhone 17",
								Description: "Phone",
								Category: "Electronics",
								Price: 80000,
								DeliveryDays: 4,
								Rating: 4.8,
								CreatedAt: time.Now(),
								UpdatedAt: time.Now(),
							},
						},
						Total: 1,
						Page: 2,
						PageSize: 10,
					}, nil
				},
			},
			checkResult: func(t *testing.T, result *domain.ProductList) {
				require.NotNil(t, result)
				require.Len(t, result.Items, 1)

				require.Equal(t, int64(1), result.Total)
				require.Equal(t, 2, result.Page)
				require.Equal(t, 10, result.PageSize)

				require.Equal(t, "iPhone 17", result.Items[0].Name)
			},
		},
		{
			name: "default page",

			filter: domain.ListFilter{
				Page: 0,
				PageSize: 10,
			},

			mock: &MockRepository{
				ListFunc: func(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
					require.Equal(t, 1, filter.Page)
					require.Equal(t, 10, filter.PageSize)

					return &domain.ProductList{}, nil
				},
			},
		},
		{
			name: "default page size",

			filter: domain.ListFilter{
				Page: 1,
				PageSize: 0,
			},

			mock: &MockRepository{
				ListFunc: func(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
					require.Equal(t, 1, filter.Page)
					require.Equal(t, 20, filter.PageSize)

					return &domain.ProductList{}, nil
				},
			},
		},
		{
			name: "default page and page size",

			filter: domain.ListFilter{
				Page: 0,
				PageSize: 0,
			},

			mock: &MockRepository{
				ListFunc: func(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
					require.Equal(t, 1, filter.Page)
					require.Equal(t, 20, filter.PageSize)

					return &domain.ProductList{}, nil
				},
			},
		},
		{
			name: "repository error",

			filter: domain.ListFilter{
				Page: 1,
				PageSize: 20,
			},

			mock: &MockRepository{
				ListFunc: func(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
					return nil, repoErr
				},
			},

			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock)

			result, err := svc.List(context.Background(), tt.filter)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)

				require.Equal(t, 1, tt.mock.ListCalls)

				require.Zero(t, tt.mock.CreateCalls)
				require.Zero(t, tt.mock.GetByIDCalls)
				require.Zero(t, tt.mock.UpdateCalls)
				require.Zero(t, tt.mock.DeleteCalls)

				return
			}

			require.NoError(t, err)

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			require.Equal(t, 1, tt.mock.ListCalls)

			require.Zero(t, tt.mock.CreateCalls)
			require.Zero(t, tt.mock.GetByIDCalls)
			require.Zero(t, tt.mock.UpdateCalls)
			require.Zero(t, tt.mock.DeleteCalls)
		})
	}
}

func TestService_Update(t *testing.T) {
	validID := uuid.New()

	t.Parallel()

	now := time.Now()

	repoErr := errors.New("repository error")

	product := &domain.Product{
		ID: validID,
		Name: "iPhone 17 Pro",
		Description: "Best phone",
		Category: "Electronics",
		Price: 120000,
		DeliveryDays: 2,
		Rating: 4.9,
		CreatedAt: now,
		UpdatedAt: now,
	}	

	tests := []struct {
		name string
		id uuid.UUID
		input *UpdateProduct
		mock *MockRepository
		wantErr error
		checkResult func(t *testing.T, p *domain.Product)
	}{
		{
			name: "success",

			id: validID,

			input: &UpdateProduct{
				Name: strPtr("iPhone 17 Pro Max"),
				Price: int64Ptr(150000),
			},

			mock: &MockRepository{
				GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
					return product, nil
				},
				UpdateFunc: func(ctx context.Context, p *domain.Product) error {
					return nil
				},
			},

			checkResult: func(t *testing.T, p *domain.Product) {
				require.NotNil(t, p)

				require.Equal(t, "iPhone 17 Pro Max", p.Name)
				require.Equal(t, int64(150000), p.Price)

				require.Equal(t, "Best phone", p.Description)
				require.Equal(t, "Electronics", p.Category)
				require.Equal(t, 2, p.DeliveryDays)
				require.Equal(t, 4.9, p.Rating)

				require.True(t, p.UpdatedAt.After(now))

				require.Equal(t, validID, p.ID)
				require.Equal(t, now, p.CreatedAt)
			},
		},
		{
			name: "invalid id",
			id: uuid.Nil,
			input: &UpdateProduct{Name: strPtr("test")},
			mock: &MockRepository{},
			wantErr: ErrInvalidID,
		},
		{
			name: "nil input",
			id: validID,
			input: nil,
			mock: &MockRepository{},
			wantErr: ErrNilInput,
		},
		{
			name: "empty patch",
			id: validID,
			input: &UpdateProduct{
				Name: nil,
				Description: nil,
				Category: nil,
				Price: nil,
				DeliveryDays: nil,
			},

			mock: &MockRepository{
				GetByIDFunc: func(ctx context.Context, u uuid.UUID) (*domain.Product, error) {
					return product, nil
				},
			},

			wantErr: ErrEmptyUpdate,
		},
		{
			name: "product not found",
			id: validID,
			input: &UpdateProduct{Name: strPtr("test")},
			mock: &MockRepository{
				GetByIDFunc: func(ctx context.Context, u uuid.UUID) (*domain.Product, error) {
					return nil, domain.ErrProductNotFound
				},
			},
			wantErr: domain.ErrProductNotFound,
		},
		{
			name: "repository error",
			id: validID,
			input: &UpdateProduct{Name: strPtr("test")},
			mock: &MockRepository{
				GetByIDFunc: func(ctx context.Context, u uuid.UUID) (*domain.Product, error) {
					return product, nil
				},
				UpdateFunc: func(ctx context.Context, p *domain.Product) error {
					return repoErr
				},
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock)

			err := svc.Update(context.Background(), tt.id, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				
				if tt.id == uuid.Nil || tt.input == nil {
					require.Zero(t, tt.mock.GetByIDCalls)
				}

				if tt.name == "empty patch" {
					require.Equal(t, 1, tt.mock.GetByIDCalls)
					require.Zero(t, tt.mock.UpdateCalls)
				}

				if tt.name == "product not found" {
					require.Equal(t, 1, tt.mock.GetByIDCalls)
					require.Zero(t, tt.mock.UpdateCalls)
				}

				if tt.name == "repository error" {
					require.Equal(t, 1, tt.mock.GetByIDCalls)
					require.Equal(t, 1, tt.mock.UpdateCalls)
				}

				require.Zero(t, tt.mock.CreateCalls)
				require.Zero(t, tt.mock.DeleteCalls)
				require.Zero(t, tt.mock.ListCalls)

				return
			}

			require.NoError(t, err)

			require.Equal(t, 1, tt.mock.GetByIDCalls)
			require.Equal(t, 1, tt.mock.UpdateCalls)

			require.Zero(t, tt.mock.CreateCalls)
			require.Zero(t, tt.mock.DeleteCalls)
			require.Zero(t, tt.mock.ListCalls)
		})
	}
}