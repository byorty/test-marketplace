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
	"github.com/byorty/test-marketplace/services/order-service/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestLogger() *zap.Logger {
	return zap.NewNop()
}

func newTestService(repo *mocks.MockOrderRepository, product *mocks.MockProductClient) *OrderService {
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
		repo    *mocks.MockOrderRepository
		product *mocks.MockProductClient
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

			repo: &mocks.MockOrderRepository{
				AddToCartFn: func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
					return nil
				},
			},

			product: &mocks.MockProductClient{
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
			repo:    &mocks.MockOrderRepository{},
			product: &mocks.MockProductClient{},
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

			repo:    &mocks.MockOrderRepository{},
			product: &mocks.MockProductClient{},
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

			repo:    &mocks.MockOrderRepository{},
			product: &mocks.MockProductClient{},
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

			repo:    &mocks.MockOrderRepository{},
			product: &mocks.MockProductClient{},
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

			repo: &mocks.MockOrderRepository{},

			product: &mocks.MockProductClient{
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

			repo: &mocks.MockOrderRepository{
				AddToCartFn: func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
					return domain.ErrProductNotInCart
				},
			},

			product: &mocks.MockProductClient{
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
		repo        *mocks.MockOrderRepository
		checkResult func(t *testing.T, cart *domain.Cart)
		wantErr     error
	}{
		{
			name:   "success",
			userID: userID,
			repo: &mocks.MockOrderRepository{
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
			repo:   &mocks.MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:   "repository error",
			userID: userID,
			repo: &mocks.MockOrderRepository{
				GetCartFn: func(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error) {
					return nil, domain.ErrCartEmpty
				},
			},
			wantErr: domain.ErrCartEmpty,
		},
		{
			name:   "empty cart success",
			userID: userID,
			repo: &mocks.MockOrderRepository{
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
	cartItemID := uuid.New()

	tests := []struct {
		name       string
		userID     uuid.UUID
		cartItemID uuid.UUID
		mock       *mocks.MockOrderRepository
		wantErr    error
	}{
		{
			name:       "success",
			userID:     userID,
			cartItemID: cartItemID,
			mock: &mocks.MockOrderRepository{
				RemoveFromCartFn: func(
					ctx context.Context,
					uid, cid uuid.UUID,
				) error {
					require.Equal(t, userID, uid)
					require.Equal(t, cartItemID, cid)
					return nil
				},
			},
		},
		{
			name:       "invalid user id",
			userID:     uuid.Nil,
			cartItemID: cartItemID,
			mock:       &mocks.MockOrderRepository{},
			wantErr:    ErrInvalidUserID,
		},
		{
			name:       "invalid cart item id",
			userID:     userID,
			cartItemID: uuid.Nil,
			mock:       &mocks.MockOrderRepository{},
			wantErr:    ErrInvalidCartItemID,
		},
		{
			name:       "cart item not found",
			userID:     userID,
			cartItemID: cartItemID,
			mock: &mocks.MockOrderRepository{
				RemoveFromCartFn: func(
					ctx context.Context,
					uid, cid uuid.UUID,
				) error {
					require.Equal(t, userID, uid)
					require.Equal(t, cartItemID, cid)
					return domain.ErrCartItemNotFound
				},
			},
			wantErr: domain.ErrCartItemNotFound,
		},
		{
			name:       "repository error",
			userID:     userID,
			cartItemID: cartItemID,
			mock: &mocks.MockOrderRepository{
				RemoveFromCartFn: func(
					ctx context.Context,
					uid, cid uuid.UUID,
				) error {
					return errors.New("database error")
				},
			},
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock, nil)

			err := svc.RemoveFromCart(
				context.Background(),
				tt.userID,
				tt.cartItemID,
			)

			if tt.wantErr != nil {
				require.Error(t, err)

				if tt.name == "repository error" {
					require.ErrorContains(t, err, "remove from cart")
					require.ErrorContains(t, err, "database error")
				} else {
					require.ErrorIs(t, err, tt.wantErr)
				}

				if tt.userID != uuid.Nil && tt.cartItemID != uuid.Nil {
					require.Equal(t, 1, tt.mock.RemoveFromCartFnCalls)
				}

				return
			}

			require.NoError(t, err)
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
		mock *mocks.MockOrderRepository
		wantErr error
	}{
		{
			name: "success",
			userID: userID,

			mock: &mocks.MockOrderRepository{
				ClearCartFn: func(ctx context.Context, userID uuid.UUID) error {
					return nil
				},
			},
		},
		{
			name: "invalid user id",
			userID: uuid.Nil,
			mock: &mocks.MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "repository error",
			userID: userID,

			mock: &mocks.MockOrderRepository{
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
		mock        *mocks.MockOrderRepository
		checkResult func(t *testing.T, order *domain.Order)
		wantErr     error
	}{
		{
			name:    "success",
			userID:  userID,
			orderID: orderID,
			mock: &mocks.MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:           orderID,
						UserID:       userID, 
						Status:       domain.StatusCreated,
						TotalPrice:        1500,
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
				require.EqualValues(t, 1500, o.TotalPrice)

				require.False(t, o.CreatedAt.IsZero())
				require.False(t, o.DeliveryDate.IsZero())
			},
		},
		{
			name:    "forbidden - order belongs to another user",
			userID:  userID,
			orderID: orderID,
			mock: &mocks.MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: uuid.New(), 
						Status: domain.StatusCreated,
						TotalPrice:  1500,
					}, nil
				},
			},
			wantErr: ErrForbidden,
		},
		{
			name:    "invalid user id",
			userID:  uuid.Nil,
			orderID: orderID,
			mock:    &mocks.MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:    "invalid order id",
			userID:  userID,
			orderID: uuid.Nil,
			mock:    &mocks.MockOrderRepository{},
			wantErr: ErrInvalidID,
		},
		{
			name:    "repository error",
			userID:  userID,
			orderID: orderID,
			mock: &mocks.MockOrderRepository{
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
		mock        *mocks.MockOrderRepository
		checkResult func(t *testing.T, items []domain.OrderItem)
		wantErr     error
	}{
		{
			name:    "success",
			userID:  userID,
			orderID: orderID,
			mock: &mocks.MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: userID,
						Status: domain.StatusCreated,
						TotalPrice:  370000,
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
			mock:    &mocks.MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:    "invalid order id",
			userID:  userID,
			orderID: uuid.Nil,
			mock:    &mocks.MockOrderRepository{},
			wantErr: ErrInvalidOrderID,
		},
		{
			name:    "forbidden - order belongs to another user",
			userID:  userID,
			orderID: orderID,
			mock: &mocks.MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: uuid.New(),
						Status: domain.StatusCreated,
						TotalPrice:  370000,
					}, nil
				},
			},
			wantErr: ErrForbidden,
		},
		{
			name:    "order not found",
			userID:  userID,
			orderID: orderID,
			mock: &mocks.MockOrderRepository{
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
			mock: &mocks.MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: userID,
						Status: domain.StatusCreated,
						TotalPrice:  370000,
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

	successRepo := &mocks.MockOrderRepository{}

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

	successProduct := &mocks.MockProductClient{
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
		repo        *mocks.MockOrderRepository
		product     *mocks.MockProductClient
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
				require.EqualValues(t, 240000, o.TotalPrice)

				require.False(t, o.CreatedAt.IsZero())
				require.False(t, o.DeliveryDate.IsZero())
			},
		},
		{
			name:    "invalid user id",
			userID:  uuid.Nil,
			repo:    &mocks.MockOrderRepository{},
			product: &mocks.MockProductClient{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:   "empty cart",
			userID: userID,
			repo: &mocks.MockOrderRepository{
				GetCartFn: func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{}, nil
				},
			},
			product: &mocks.MockProductClient{},
			wantErr: domain.ErrCartEmpty,
		},
		{
			name:   "product not found",
			userID: userID,
			repo: &mocks.MockOrderRepository{
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
			product: &mocks.MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error) {
					return nil, product.ErrProductNotFound
				},
			},
			wantErr: ErrProductNotFound,
		},
		{
			name:   "transaction error",
			userID: userID,
			product: &mocks.MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error) {
					return &client.ProductResponse{
						Id:           id,
						Name:         "iPhone",
						Price:        100,
						DeliveryDays: 2,
					}, nil
				},
			},
			repo: func() *mocks.MockOrderRepository {
				repo := &mocks.MockOrderRepository{}
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