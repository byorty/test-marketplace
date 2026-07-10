package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	"github.com/byorty/test-marketplace/services/product-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func ptr(i int) *int {
	return &i
}

func TestHandler_CreateProduct(t *testing.T) {
	t.Parallel()

	srvcError := errors.New("service error")

	tests := []struct {
		name string
		req api.CreateProductRequestObject
		mock *MockService
		checkResult func(t *testing.T, resp api.CreateProductResponseObject)
	}{
		{
			name: "success",

			req: api.CreateProductRequestObject{
				Body: &api.ProductCreateRequest{
					Name: "iPhone 17 Pro",
					Description: "Best phone",
					Category: "Electronics",
					Price: 120000,
					DeliveryDays: 3,
				},
			},

			mock: &MockService{
				CreateFunc: func(ctx context.Context, input *service.CreateProduct) (*domain.Product, error) {
					return &domain.Product{
						ID: uuid.New(),
						Name: input.Name,
						Description: input.Description,
						Category: input.Category,
						Price: input.Price,
						DeliveryDays: input.DeliveryDays,
						Rating: 0,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				},
			},

			checkResult: func(t *testing.T, resp api.CreateProductResponseObject) {
				response, ok := resp.(api.CreateProduct201JSONResponse)
				require.True(t, ok)

				require.Equal(t, "iPhone 17 Pro", response.Name)
				require.Equal(t, "Best phone", response.Description)
				require.Equal(t, "Electronics", response.Category)
				require.Equal(t, int64(120000), response.Price)
				require.Equal(t, 3, response.DeliveryDays)
			},
		},
		{
			name: "validation error",

			req: api.CreateProductRequestObject{
				Body: &api.ProductCreateRequest{
					Name: "",
				},
			},

			mock: &MockService{
				CreateFunc: func(ctx context.Context, cp *service.CreateProduct) (*domain.Product, error) {
					return nil, service.ErrInvalidInput
				},
			},

			checkResult: func(t *testing.T, resp api.CreateProductResponseObject) {
				response, ok := resp.(api.CreateProduct400JSONResponse)
				require.True(t, ok)

				require.Equal(t, "validation_error", response.Code)
				require.Equal(t, service.ErrInvalidInput.Error(), response.Message)
			},
		},
		{
			name: "internal error",

			req: api.CreateProductRequestObject{
				Body: &api.ProductCreateRequest{
					Name: "iPhone",
				},
			},

			mock: &MockService{
				CreateFunc: func(ctx context.Context, cp *service.CreateProduct) (*domain.Product, error) {
					return nil, srvcError
				},
			},

			checkResult: func(t *testing.T, resp api.CreateProductResponseObject) {
				response, ok := resp.(api.CreateProduct500JSONResponse)
				require.True(t, ok)

				require.Equal(t, "internal_error", response.Code)
				require.Equal(t, http.StatusText(http.StatusInternalServerError), response.Message)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				service: tt.mock,
				log: newTestLogger(),
			}

			resp, err := handler.CreateProduct(context.Background(), tt.req)

			require.NoError(t, err)

			tt.checkResult(t, resp)

			require.Equal(t, 1, tt.mock.CreateCalls)

			require.Zero(t, tt.mock.GetByIDCalls)
			require.Zero(t, tt.mock.UpdateCalls)
			require.Zero(t, tt.mock.DeleteCalls)
			require.Zero(t, tt.mock.ListCalls)
		})
	}
}

func TestHandler_GetByID(t *testing.T) {
	t.Parallel()

	srvcErr := errors.New("service error")

	product := &domain.Product{
		ID: uuid.New(),
		Name: "iPhone 17",
		Description: "Best phone",
		Category: "Electronics",
		Price: 100000,
		DeliveryDays: 5,
		Rating: 4.8,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name string
		req api.GetByIDRequestObject
		mock *MockService
		checkResult func(t *testing.T, resp api.GetByIDResponseObject)
	}{
		{
			name: "success",

			req: api.GetByIDRequestObject{
				Id: product.ID,
			},

			mock: &MockService{
				GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
					require.Equal(t, product.ID, id)

					return product, nil
				},
			},

			checkResult: func(t *testing.T, resp api.GetByIDResponseObject) {
				response, ok := resp.(api.GetByID200JSONResponse)
				require.True(t, ok)

				require.Equal(t, product.ID, response.Id)
				require.Equal(t, product.Name, response.Name)
				require.Equal(t, product.Description, response.Description)
				require.Equal(t, product.Category, response.Category)
				require.Equal(t, product.DeliveryDays, response.DeliveryDays)
				require.Equal(t, product.Price, response.Price)
				require.Equal(t, product.Rating, response.Rating)
			},
		},
		{
			name: "product not found",

			req: api.GetByIDRequestObject{
				Id: uuid.New(),
			},

			mock: &MockService{
				GetByIDFunc: func(ctx context.Context, u uuid.UUID) (*domain.Product, error) {
					return nil, domain.ErrProductNotFound
				},
			},

			checkResult: func(t *testing.T, resp api.GetByIDResponseObject) {
				response, ok := resp.(api.GetByID404JSONResponse)
				require.True(t, ok)

				require.Equal(t, "product_not_found", response.Code)
				require.Equal(t, domain.ErrProductNotFound.Error(), response.Message)
			},
		},
		{
			name: "internal error",

			req: api.GetByIDRequestObject{
				Id: uuid.New(),
			},

			mock: &MockService{
				GetByIDFunc: func(ctx context.Context, u uuid.UUID) (*domain.Product, error) {
					return nil, srvcErr
				},
			},

			checkResult: func(t *testing.T, resp api.GetByIDResponseObject) {
				response, ok := resp.(api.GetByID500JSONResponse)
				require.True(t, ok)

				require.Equal(t, "internal_error", response.Code)
				require.Equal(t, http.StatusText(http.StatusInternalServerError), response.Message)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				service: tt.mock,
				log: newTestLogger(),
			}

			resp, err := handler.GetByID(context.Background(), tt.req)

			require.NoError(t, err)

			tt.checkResult(t, resp)

			require.Equal(t, 1, tt.mock.GetByIDCalls)

			require.Zero(t, tt.mock.CreateCalls)
			require.Zero(t, tt.mock.UpdateCalls)
			require.Zero(t, tt.mock.DeleteCalls)
			require.Zero(t, tt.mock.ListCalls)
		})
	}
}

func TestHandler_DeleteProduct(t *testing.T) {
	requestID := uuid.New()

	t.Parallel()

	srvcErr := errors.New("service error")

	tests := []struct {
		name string
		req api.DeleteProductRequestObject
		mock *MockService
		checkResult func(t *testing.T, resp api.DeleteProductResponseObject)
	}{
		{
			name: "success",

			req: api.DeleteProductRequestObject{
				Id: uuid.New(),
			},

			mock: &MockService{
				DeleteFunc: func(ctx context.Context, u uuid.UUID) error {
					return nil
				},
			},

			checkResult: func(t *testing.T, resp api.DeleteProductResponseObject) {
				_, ok := resp.(api.DeleteProduct204Response)
				require.True(t, ok)
			},
		},
		{
			name: "product not found",

			req: api.DeleteProductRequestObject{
				Id: requestID,
			},

			mock: &MockService{
				DeleteFunc: func(ctx context.Context, u uuid.UUID) error {
					return domain.ErrProductNotFound
				},
			},

			checkResult: func(t *testing.T, resp api.DeleteProductResponseObject) {
				response, ok := resp.(api.DeleteProduct404JSONResponse)
				require.True(t, ok)

				require.Equal(t, "product_not_found", response.Code)
				require.Equal(t, domain.ErrProductNotFound.Error(), response.Message)
			},
		},
		{
			name: "internal error",

			req: api.DeleteProductRequestObject{
				Id: uuid.New(),
			},

			mock: &MockService{
				DeleteFunc: func(ctx context.Context, u uuid.UUID) error {
					return srvcErr
				},
			},

			checkResult: func(t *testing.T, resp api.DeleteProductResponseObject) {
				response, ok := resp.(api.DeleteProduct500JSONResponse)
				require.True(t, ok)

				require.Equal(t, "internal_error", response.Code)
				require.Equal(t, http.StatusText(http.StatusInternalServerError), response.Message)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				service: tt.mock,
				log: newTestLogger(),
			}

			resp, err := handler.DeleteProduct(context.Background(), tt.req)

			require.NoError(t, err)
			
			tt.checkResult(t, resp)

			require.Equal(t, 1, tt.mock.DeleteCalls)

			require.Zero(t, tt.mock.CreateCalls)
			require.Zero(t, tt.mock.GetByIDCalls)
			require.Zero(t, tt.mock.UpdateCalls)
			require.Zero(t, tt.mock.ListCalls)
		})
	}
}

func TestHandler_UpdateProduct(t *testing.T) {
	requestID := uuid.New()

	t.Parallel()

	srvcErr := errors.New("service error")

	name := "iPhone 17"
	price := int64(100000)

	tests := []struct {
		name string
		req api.UpdateProductRequestObject
		mock *MockService
		checkResult func(t *testing.T, resp api.UpdateProductResponseObject)
	}{
		{
			name: "success",

			req: api.UpdateProductRequestObject{
				Id: requestID,
				Body: &api.ProductUpdateRequest{
					Name: &name,
					Price: &price,
				},
			},

			mock: &MockService{
				UpdateFunc: func(ctx context.Context, id uuid.UUID, input *service.UpdateProduct) error {
					require.NotNil(t, input)
					require.Equal(t, name, *input.Name)
					require.Equal(t, price, *input.Price)

					return nil
				},
			},

			checkResult: func(t *testing.T, resp api.UpdateProductResponseObject) {
				_, ok := resp.(api.UpdateProduct200JSONResponse)
				require.True(t, ok)
			},
		},
		{
			name: "validation error",

			req: api.UpdateProductRequestObject{
				Id: uuid.New(),
				Body: &api.ProductUpdateRequest{},
			},

			mock: &MockService{
				UpdateFunc: func(ctx context.Context, u uuid.UUID, up *service.UpdateProduct) error {
					return service.ErrInvalidInput
				},
			},

			checkResult: func(t *testing.T, resp api.UpdateProductResponseObject) {
				response, ok := resp.(api.UpdateProduct400JSONResponse)
				require.True(t, ok)

				require.Equal(t, "validation_error", response.Code)
				require.Equal(t, service.ErrInvalidInput.Error(), response.Message)
			},
		},
		{
			name: "empty update",

			req: api.UpdateProductRequestObject{
				Id: uuid.New(),
				Body: &api.ProductUpdateRequest{},
			},

			mock: &MockService{
				UpdateFunc: func(ctx context.Context, id uuid.UUID, input *service.UpdateProduct) error {
					return service.ErrEmptyUpdate
				},
			},

			checkResult: func(t *testing.T, resp api.UpdateProductResponseObject) {
				response, ok := resp.(api.UpdateProduct400JSONResponse)
				require.True(t, ok)

				require.Equal(t, "empty_update", response.Code)
				require.Equal(t, service.ErrEmptyUpdate.Error(), response.Message)
			},
		},
		{
			name: "product not found",

			req: api.UpdateProductRequestObject{
				Id: uuid.New(),
				Body: &api.ProductUpdateRequest{},
			},

			mock: &MockService{
				UpdateFunc: func(ctx context.Context, u uuid.UUID, up *service.UpdateProduct) error {
					return domain.ErrProductNotFound
				},
			},

			checkResult: func(t *testing.T, resp api.UpdateProductResponseObject) {
				response, ok := resp.(api.UpdateProduct404JSONResponse)
				require.True(t, ok)

				require.Equal(t, "product_not_found", response.Code)
				require.Equal(t, domain.ErrProductNotFound.Error(), response.Message)
			},
		},
		{
			name: "internal error",

			req: api.UpdateProductRequestObject{
				Id: uuid.New(),
				Body: &api.ProductUpdateRequest{},
			},

			mock: &MockService{
				UpdateFunc: func(ctx context.Context, u uuid.UUID, up *service.UpdateProduct) error {
					return srvcErr
				},
			},

			checkResult: func(t *testing.T, resp api.UpdateProductResponseObject) {
				response, ok := resp.(api.UpdateProduct500JSONResponse)
				require.True(t, ok)

				require.Equal(t, "internal_error", response.Code)
				require.Equal(t, http.StatusText(http.StatusInternalServerError), response.Message)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				service: tt.mock,
				log: newTestLogger(),
			}

			resp, err := handler.UpdateProduct(context.Background(), tt.req)

			require.NoError(t, err)

			tt.checkResult(t, resp)

			require.Equal(t, 1, tt.mock.UpdateCalls)

			require.Zero(t, tt.mock.CreateCalls)
			require.Zero(t, tt.mock.GetByIDCalls)
			require.Zero(t, tt.mock.DeleteCalls)
			require.Zero(t, tt.mock.ListCalls)
		})
	}
}

func TestHandler_GetProducts(t *testing.T) {
	t.Parallel()

	srvcErr := errors.New("service error")

	product := &domain.Product{
		ID: uuid.New(),
		Name: "iPhone 17",
		Description: "Best phone",
		Category: "Electronics",
		Price: 100000,
		DeliveryDays: 3,
		Rating: 4.8,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	tests := []struct {
		name string
		req api.GetProductsRequestObject
		mock *MockService
		checkResult func(t *testing.T, resp api.GetProductsResponseObject)
	}{
		{
			name: "success",

			req: api.GetProductsRequestObject{
				Params: api.GetProductsParams{
					Page: ptr(1),
					PageSize: ptr(10),
				},
			},

			mock: &MockService{
				ListFunc: func(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
					require.Equal(t, 1, filter.Page)
					require.Equal(t, 10, filter.PageSize) 

					return &domain.ProductList{
						Items: []*domain.Product{product},
						Total: 1,
						Page: 1,
						PageSize: 10,
					}, nil
				},
			},

			checkResult: func(t *testing.T, resp api.GetProductsResponseObject) {
				response, ok := resp.(api.GetProducts200JSONResponse)
				require.True(t, ok)

				require.Equal(t, int64(1), response.Total)
				require.Equal(t, 1, response.Page)
				require.Equal(t, 10, response.PageSize)

				require.Len(t, response.Items, 1)

				require.Equal(t, product.ID, response.Items[0].Id)
				require.Equal(t, product.Name, response.Items[0].Name)
				require.Equal(t, product.Description, response.Items[0].Description)
				require.Equal(t, product.Category, response.Items[0].Category)
				require.Equal(t, product.Price, response.Items[0].Price)
			},
		},
		{
			name: "validation error",

			req: api.GetProductsRequestObject{},

			mock: &MockService{
				ListFunc: func(ctx context.Context, lf domain.ListFilter) (*domain.ProductList, error) {
					return nil, service.ErrInvalidInput
				},
			},

			checkResult: func(t *testing.T, resp api.GetProductsResponseObject) {
				response, ok := resp.(api.GetProducts400JSONResponse)
				require.True(t, ok)

				require.Equal(t, "validation_error", response.Code)
				require.Equal(t, service.ErrInvalidInput.Error(), response.Message)
			},
		},
		{
			name: "internal error",

			req: api.GetProductsRequestObject{},

			mock: &MockService{
				ListFunc: func(ctx context.Context, lf domain.ListFilter) (*domain.ProductList, error) {
					return nil, srvcErr
				},
			},

			checkResult: func(t *testing.T, resp api.GetProductsResponseObject) {
				response, ok := resp.(api.GetProducts500JSONResponse)
				require.True(t, ok)

				require.Equal(t, "internal_error", response.Code)
				require.Equal(t, http.StatusText(http.StatusInternalServerError), response.Message)
			},
		},
	}

	for _, tt := range tests {
		tt := tt 

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				service: tt.mock,
				log: newTestLogger(),
			}

			resp, err := handler.GetProducts(context.Background(), tt.req)

			require.NoError(t, err)

			tt.checkResult(t, resp)

			require.Equal(t, 1, tt.mock.ListCalls)

			require.Zero(t, tt.mock.CreateCalls)
			require.Zero(t, tt.mock.GetByIDCalls)
			require.Zero(t, tt.mock.UpdateCalls)
			require.Zero(t, tt.mock.DeleteCalls)
		})
	}
}