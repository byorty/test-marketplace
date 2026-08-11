package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/byorty/test-marketplace/services/common/client/product"
	client "github.com/byorty/test-marketplace/services/common/client/product/generated"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestLogger() *zap.Logger {
	return zap.NewNop()
}

func newTestService(repo *MockOrderRepository, product *MockProductClient) *OrderService {
	return &OrderService{
		repo: repo,
		productClient: product,
		log: newTestLogger(),
	}
}

func TestService_AddToCart(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name    string
		userID  uuid.UUID
		input   *domain.CartItem
		repo    *MockOrderRepository
		product *MockProductClient
		wantErr error
	}{
		{
			name:   "success",
			userID: userID,

			input: &domain.CartItem{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  2,
			},

			repo: &MockOrderRepository{
				AddToCartFn: func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
					return nil
				},
			},

			product: &MockProductClient{
				GetProductFn: func(
					ctx context.Context,
					id uuid.UUID,
				) (*client.ProductResponse, error) {
					return &client.ProductResponse{
						Id: id,
					}, nil
				},
			},
		},
		{
			name:    "nil input",
			userID:  userID,
			input:   nil,
			repo:    &MockOrderRepository{},
			product: &MockProductClient{},
			wantErr: ErrInvalidInput,
		},
		{
			name:   "invalid user id",
			userID: uuid.Nil,

			input: &domain.CartItem{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  2,
			},

			repo:    &MockOrderRepository{},
			product: &MockProductClient{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:   "invalid product id",
			userID: userID,

			input: &domain.CartItem{
				ID:        uuid.New(),
				ProductID: uuid.Nil,
				Quantity:  2,
			},

			repo:    &MockOrderRepository{},
			product: &MockProductClient{},
			wantErr: ErrInvalidProductID,
		},
		{
			name:   "invalid quantity",
			userID: userID,

			input: &domain.CartItem{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  0,
			},

			repo:    &MockOrderRepository{},
			product: &MockProductClient{},
			wantErr: ErrInvalidQuantity,
		},
		{
			name:   "product not found",
			userID: userID,

			input: &domain.CartItem{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  2,
			},

			repo: &MockOrderRepository{},

			product: &MockProductClient{
				GetProductFn: func(
					ctx context.Context,
					id uuid.UUID,
				) (*client.ProductResponse, error) {
					return nil, product.ErrProductNotFound
				},
			},

			wantErr: ErrProductNotFound,
		},
		{
			name:   "repository error",
			userID: userID,

			input: &domain.CartItem{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  2,
			},

			repo: &MockOrderRepository{
				AddToCartFn: func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
					return domain.ErrProductNotInCart
				},
			},

			product: &MockProductClient{
				GetProductFn: func(
					ctx context.Context,
					id uuid.UUID,
				) (*client.ProductResponse, error) {
					return &client.ProductResponse{Id: id}, nil
				},
			},

			wantErr: domain.ErrProductNotInCart,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, tt.product) 

			err := svc.AddToCart(
				context.Background(),
				tt.userID,
				tt.input,
			)

			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.wantErr != nil {
					require.ErrorIs(t, err, tt.wantErr)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, 1, tt.repo.AddToCartFnCalls)
			require.Equal(t, 1, tt.product.GetProductFnCalls)
		})
	}
}
func TestService_GetCart(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	productID1 := uuid.New()
	productID2 := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		repo        *MockOrderRepository
		checkResult func(t *testing.T, cart *domain.Cart)
		wantErr     error
	}{
		{
			name:   "success",
			userID: userID,
			repo: &MockOrderRepository{
				GetCartFn: func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{
						{
							ID:        uuid.New(),
							UserID:    id,
							ProductID: productID1,
							Quantity:  2,
						},
						{
							ID:        uuid.New(),
							UserID:    id,
							ProductID: productID2,
							Quantity:  5,
						},
					}, nil
				},
			},
			checkResult: func(t *testing.T, cart *domain.Cart) {
				require.NotNil(t, cart)
				require.Len(t, cart.Items, 2)

				require.Equal(t, 2, cart.Items[0].Quantity)
				require.Equal(t, 5, cart.Items[1].Quantity)

				require.Equal(t, userID, cart.Items[0].UserID)
				require.Equal(t, userID, cart.Items[1].UserID)

				require.Equal(t, productID1, cart.Items[0].ProductID)
				require.Equal(t, productID2, cart.Items[1].ProductID)
			},
		},
		{
			name:   "invalid user id",
			userID: uuid.Nil,
			repo:   &MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:   "repository error",
			userID: userID,
			repo: &MockOrderRepository{
				GetCartFn: func(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error) {
					return nil, domain.ErrCartEmpty
				},
			},
			wantErr: domain.ErrCartEmpty,
		},
		{
			name:   "empty cart success",
			userID: userID,
			repo: &MockOrderRepository{
				GetCartFn: func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{}, nil
				},
			},
			checkResult: func(t *testing.T, cart *domain.Cart) {
				require.NotNil(t, cart)
				require.Empty(t, cart.Items)
				require.Equal(t, int64(0), cart.TotalPrice)
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, nil)

			cart, err := svc.GetCart(context.Background(), tt.userID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, cart)

				if tt.userID != uuid.Nil {
					require.Equal(t, 1, tt.repo.GetCartFnCalls)
				}
				return
			}

			require.NoError(t, err)
			tt.checkResult(t, cart)

			require.Equal(t, 1, tt.repo.GetCartFnCalls)
		})
	}
}
func TestService_RemoveFromCart(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	productId := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		productID uuid.UUID
		mock      *MockOrderRepository
		wantErr   error
	}{
		{
			name:      "success",
			userID:    userID,
			productID: productId,
			mock: &MockOrderRepository{

				GetCartItemFn: func(ctx context.Context, uid, pid uuid.UUID) (*domain.CartItem, error) {
					require.Equal(t, userID, uid)
					require.Equal(t, productId, pid)
					return &domain.CartItem{
						ID:        uuid.New(),
						UserID:    uid,
						ProductID: pid,
						Quantity:  2,
					}, nil
				},
				RemoveFromCartFn: func(ctx context.Context, uid, pid uuid.UUID) error {
					require.Equal(t, userID, uid)
					require.Equal(t, productId, pid)
					return nil
				},
			},
		},
		{
			name:      "invalid user id",
			userID:    uuid.Nil,
			productID: productId,
			mock:      &MockOrderRepository{},
			wantErr:   ErrInvalidUserID,
		},
		{
			name:      "invalid product id",
			userID:    userID,
			productID: uuid.Nil,
			mock:      &MockOrderRepository{},
			wantErr:   ErrInvalidProductID,
		},
		{
			name:      "product not in cart",
			userID:    userID,
			productID: productId,
			mock: &MockOrderRepository{
				GetCartItemFn: func(ctx context.Context, uid, pid uuid.UUID) (*domain.CartItem, error) {
					return nil, domain.ErrCartItemNotFound
				},
			},
			wantErr: domain.ErrCartItemNotFound,
		},
		{
			name:      "repository error on remove",
			userID:    userID,
			productID: productId,
			mock: &MockOrderRepository{
				GetCartItemFn: func(ctx context.Context, uid, pid uuid.UUID) (*domain.CartItem, error) {
					return &domain.CartItem{
						ID:        uuid.New(),
						UserID:    uid,
						ProductID: pid,
						Quantity:  2,
					}, nil
				},
				RemoveFromCartFn: func(ctx context.Context, uid, pid uuid.UUID) error {
					return domain.ErrCartItemNotFound
				},
			},
			wantErr: domain.ErrCartItemNotFound,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock, nil)

			err := svc.RemoveFromCart(context.Background(), tt.userID, tt.productID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)

				if tt.userID != uuid.Nil && tt.productID != uuid.Nil {
					require.GreaterOrEqual(t, tt.mock.GetCartItemFnCalls, 1)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, 1, tt.mock.GetCartItemFnCalls)
			require.Equal(t, 1, tt.mock.RemoveFromCartFnCalls)
		})
	}
}

func TestService_ClearCart(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name string
		userID uuid.UUID
		mock *MockOrderRepository
		wantErr error
	}{
		{
			name: "success",
			userID: userID,

			mock: &MockOrderRepository{
				ClearCartFn: func(ctx context.Context, userID uuid.UUID) error {
					return nil
				},
			},
		},
		{
			name: "invalid user id",
			userID: uuid.Nil,
			mock: &MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "repository error",
			userID: userID,

			mock: &MockOrderRepository{
				ClearCartFn: func(ctx context.Context, userID uuid.UUID) error {
					return domain.ErrCartEmpty
				},
			},

			wantErr: domain.ErrCartEmpty,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock, nil)

			err := svc.ClearCart(context.Background(), tt.userID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				if tt.userID != uuid.Nil {
					require.Equal(t, 1, tt.mock.ClearCartFnCalls)
				}

				return
			}

			require.NoError(t, err)
			require.Equal(t, 1, tt.mock.ClearCartFnCalls)
		})
	}
}

func TestService_GetOrderByID(t *testing.T) {
	t.Parallel()

	orderID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID 
		orderID     uuid.UUID
		mock        *MockOrderRepository
		checkResult func(t *testing.T, order *domain.Order)
		wantErr     error
	}{
		{
			name:    "success",
			userID:  userID,
			orderID: orderID,
			mock: &MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:           orderID,
						UserID:       userID, 
						Status:       domain.StatusCreated,
						Total:        1500,
						CreatedAt:    time.Now(),
						DeliveryDate: time.Now().Add(48 * time.Hour),
					}, nil
				},
			},
			checkResult: func(t *testing.T, o *domain.Order) {
				require.NotNil(t, o)

				require.Equal(t, orderID, o.ID)
				require.Equal(t, userID, o.UserID)
				require.Equal(t, domain.StatusCreated, o.Status)
				require.EqualValues(t, 1500, o.Total)

				require.False(t, o.CreatedAt.IsZero())
				require.False(t, o.DeliveryDate.IsZero())
			},
		},
		{
			name:    "forbidden - order belongs to another user",
			userID:  userID,
			orderID: orderID,
			mock: &MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: uuid.New(), 
						Status: domain.StatusCreated,
						Total:  1500,
					}, nil
				},
			},
			wantErr: ErrForbidden,
		},
		{
			name:    "invalid user id",
			userID:  uuid.Nil,
			orderID: orderID,
			mock:    &MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:    "invalid order id",
			userID:  userID,
			orderID: uuid.Nil,
			mock:    &MockOrderRepository{},
			wantErr: ErrInvalidID,
		},
		{
			name:    "repository error",
			userID:  userID,
			orderID: orderID,
			mock: &MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return nil, domain.ErrOrderNotFound
				},
			},
			wantErr: domain.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock, nil)

			o, err := svc.GetOrderByID(context.Background(), tt.userID, tt.orderID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, o)

				if tt.orderID != uuid.Nil && tt.userID != uuid.Nil {
					require.Equal(t, 1, tt.mock.GetOrderByIDFnCalls)
				}
				return
			}

			require.NoError(t, err)
			tt.checkResult(t, o)
			require.Equal(t, 1, tt.mock.GetOrderByIDFnCalls)
		})
	}
}

func TestService_GetOrderItems(t *testing.T) {
	t.Parallel()

	orderID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		orderID     uuid.UUID
		mock        *MockOrderRepository
		checkResult func(t *testing.T, items []domain.OrderItem)
		wantErr     error
	}{
		{
			name:    "success",
			userID:  userID,
			orderID: orderID,
			mock: &MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: userID,
						Status: domain.StatusCreated,
						Total:  370000,
					}, nil
				},
				GetOrderItemsFn: func(ctx context.Context, id uuid.UUID) ([]domain.OrderItem, error) {
					return []domain.OrderItem{
						{
							ID:           uuid.New(),
							OrderID:      id,
							ProductID:    uuid.New(),
							ProductPrice: 120000,
							Quantity:     2,
						},
						{
							ID:           uuid.New(),
							OrderID:      id,
							ProductID:    uuid.New(),
							ProductPrice: 250000,
							Quantity:     1,
						},
					}, nil
				},
			},
			checkResult: func(t *testing.T, items []domain.OrderItem) {
				require.Len(t, items, 2)

				require.Equal(t, orderID, items[0].OrderID)
				require.Equal(t, orderID, items[1].OrderID)

				require.EqualValues(t, 120000, items[0].ProductPrice)
				require.Equal(t, 2, items[0].Quantity)

				require.EqualValues(t, 250000, items[1].ProductPrice)
				require.Equal(t, 1, items[1].Quantity)
			},
		},
		{
			name:    "invalid user id",
			userID:  uuid.Nil,
			orderID: orderID,
			mock:    &MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:    "invalid order id",
			userID:  userID,
			orderID: uuid.Nil,
			mock:    &MockOrderRepository{},
			wantErr: ErrInvalidOrderID,
		},
		{
			name:    "forbidden - order belongs to another user",
			userID:  userID,
			orderID: orderID,
			mock: &MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: uuid.New(),
						Status: domain.StatusCreated,
						Total:  370000,
					}, nil
				},
			},
			wantErr: ErrForbidden,
		},
		{
			name:    "order not found",
			userID:  userID,
			orderID: orderID,
			mock: &MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return nil, domain.ErrOrderNotFound
				},
			},
			wantErr: domain.ErrOrderNotFound,
		},
		{
			name:    "get order items error",
			userID:  userID,
			orderID: orderID,
			mock: &MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: userID,
						Status: domain.StatusCreated,
						Total:  370000,
					}, nil
				},
				GetOrderItemsFn: func(ctx context.Context, id uuid.UUID) ([]domain.OrderItem, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock, nil)

			items, err := svc.GetOrderItems(
				context.Background(),
				tt.userID,
				tt.orderID,
			)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErr) || strings.Contains(err.Error(), tt.wantErr.Error()))
				require.Nil(t, items)

				if tt.userID != uuid.Nil && tt.orderID != uuid.Nil {
					require.GreaterOrEqual(t, tt.mock.GetOrderByIDFnCalls, 1)
				}

				return
			}

			require.NoError(t, err)
			tt.checkResult(t, items)
			require.Equal(t, 1, tt.mock.GetOrderByIDFnCalls)
			require.Equal(t, 1, tt.mock.GetOrderItemsFnCalls)
		})
	}
}

func TestService_CreateOrder(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	productID := uuid.New()

	successRepo := &MockOrderRepository{}

	successRepo.Self = successRepo

	successRepo.GetCartFn = func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
		return []domain.CartItem{
			{
				ID:        uuid.New(),
				UserID:    id,
				ProductID: productID,
				Quantity:  2,
			},
		}, nil
	}

	successRepo.CreateOrderFn = func(ctx context.Context, o *domain.Order) error {
		return nil
	}

	successRepo.CreateOrderItemsFn = func(ctx context.Context, items []domain.OrderItem) error {
		return nil
	}

	successRepo.ClearCartFn = func(ctx context.Context, id uuid.UUID) error {
		return nil
	}

	successRepo.TransactionFn = func(ctx context.Context, fn func(domain.OrderRepository) error) error {
		return fn(successRepo.Self)
	}

	successProduct := &MockProductClient{
		GetProductFn: func(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error) {
			return &client.ProductResponse{
				Id:           id,
				Name:         "iPhone",
				Price:        120000,
				DeliveryDays: 3,
			}, nil
		},
	}

	tests := []struct {
		name        string
		userID      uuid.UUID
		repo        *MockOrderRepository
		product     *MockProductClient
		checkResult func(t *testing.T, o *domain.Order)
		wantErr     error
	}{
		{
			name:    "success",
			userID:  userID,
			repo:    successRepo,
			product: successProduct,
			checkResult: func(t *testing.T, o *domain.Order) {
				require.NotNil(t, o)

				require.NotEqual(t, uuid.Nil, o.ID)
				require.Equal(t, userID, o.UserID)

				require.Equal(t, domain.StatusCreated, o.Status)
				require.EqualValues(t, 240000, o.Total)

				require.False(t, o.CreatedAt.IsZero())
				require.False(t, o.DeliveryDate.IsZero())
			},
		},
		{
			name:    "invalid user id",
			userID:  uuid.Nil,
			repo:    &MockOrderRepository{},
			product: &MockProductClient{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:   "empty cart",
			userID: userID,
			repo: &MockOrderRepository{
				GetCartFn: func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{}, nil
				},
			},
			product: &MockProductClient{},
			wantErr: domain.ErrCartEmpty,
		},
		{
			name:   "product not found",
			userID: userID,
			repo: &MockOrderRepository{
				GetCartFn: func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{
						{
							ID:        uuid.New(),
							UserID:    id,
							ProductID: productID,
							Quantity:  1,
						},
					}, nil
				},
			},
			product: &MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error) {
					return nil, product.ErrProductNotFound
				},
			},
			wantErr: ErrProductNotFound,
		},
		{
			name:   "transaction error",
			userID: userID,
			product: &MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error) {
					return &client.ProductResponse{
						Id:           id,
						Name:         "iPhone",
						Price:        100,
						DeliveryDays: 2,
					}, nil
				},
			},
			repo: func() *MockOrderRepository {
				repo := &MockOrderRepository{}
				repo.Self = repo

				repo.GetCartFn = func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{
						{
							ID:        uuid.New(),
							UserID:    id,
							ProductID: productID,
							Quantity:  1,
						},
					}, nil
				}

				repo.TransactionFn = func(ctx context.Context, fn func(domain.OrderRepository) error) error {
					return domain.ErrOrderNotFound
				}

				return repo
			}(),
			wantErr: domain.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, tt.product)

			o, err := svc.CreateOrder(context.Background(), tt.userID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, o)
				return
			}

			require.NoError(t, err)

			tt.checkResult(t, o)

			require.Equal(t, 1, tt.repo.GetCartFnCalls)
			require.Equal(t, 1, tt.product.GetProductFnCalls)
			require.Equal(t, 1, tt.repo.TransactionFnCalls)
			require.Equal(t, 1, tt.repo.CreateOrderFnCalls)
			require.Equal(t, 1, tt.repo.CreateOrderItemsFnCalls)
			require.Equal(t, 1, tt.repo.ClearCartFnCalls)
		})
	}
}