package service

import (
	"context"
	"testing"
	"time"

	"github.com/byorty/test-marketplace/services/order-service/internal/client/product"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestLogger() *zap.Logger {
	return zap.NewNop()
}

func newTestService(repo *domain.MockOrderRepository, product *domain.MockProductClient) *OrderService {
	return &OrderService{
		repo: repo,
		productClient: product,
		log: newTestLogger(),
	}
}

func TestService_AddToCart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input *domain.CartItem
		repo *domain.MockOrderRepository
		product *domain.MockProductClient
		wantErr error
	}{
		{
			name: "success",

			input: &domain.CartItem{
				ID: uuid.New(),
				UserID: uuid.New(),
				ProductID: uuid.New(),
				Quantity: 2,
			},

			repo: &domain.MockOrderRepository{
				AddToCartFn: func(ctx context.Context, item *domain.CartItem) error {
					return nil
				},
			},

			product: &domain.MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*product.Product, error) {
					return &product.Product{
						ID: id,
					}, nil
				},
			},
		},
		{
			name: "nil input",

			input: nil,

			repo: &domain.MockOrderRepository{},
			
			product: &domain.MockProductClient{},

			wantErr: ErrInvalidInput,
		},
		{
			name: "invalid user id",

			input: &domain.CartItem{
				ID: uuid.New(),
				UserID: uuid.Nil,
				ProductID: uuid.New(),
				Quantity: 2,
			},

			repo: &domain.MockOrderRepository{},

			product: &domain.MockProductClient{},

			wantErr: ErrInvalidUserID,
		},
		{
			name: "invalid product id",

			input: &domain.CartItem{
				ID: uuid.New(),
				UserID: uuid.New(),
				ProductID: uuid.Nil,
				Quantity: 2,
			},

			repo: &domain.MockOrderRepository{},
			
			product: &domain.MockProductClient{},

			wantErr: ErrInvalidProductID,
		},
		{
			name: "invalid quantity",

			input: &domain.CartItem{
				ID: uuid.New(),
				UserID: uuid.New(),
				ProductID: uuid.New(),
				Quantity: 0,
			},

			repo: &domain.MockOrderRepository{},

			product: &domain.MockProductClient{},

			wantErr: ErrInvalidQuantity,
		},
		{
			name: "product not found",

			input: &domain.CartItem{
				ID: uuid.New(),
				UserID: uuid.New(),
				ProductID: uuid.New(),
				Quantity: 2,
			},

			repo: &domain.MockOrderRepository{},

			product: &domain.MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*product.Product, error) {
					return nil, product.ErrProductNotFound
				},
			},

			wantErr: ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, tt.product)

			err := svc.AddToCart(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
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
		name string
		userID uuid.UUID
		repo *domain.MockOrderRepository
		product *domain.MockProductClient
		checkResult func(t *testing.T, cart *domain.Cart)
		wantErr error
	}{
		{
			name: "success",

			userID: userID,

			repo: &domain.MockOrderRepository{
				GetCartFn: func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{
						{	
							ID: uuid.New(),
							UserID: id,
							ProductID: productID1,
							Quantity: 2,
						},
						{
							ID: uuid.New(),
							UserID: id,
							ProductID: productID2,
							Quantity: 5,
						},
					}, nil
				}, 
			},

			product: &domain.MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*product.Product, error) {
					switch id {
					case productID1:
						return &product.Product{
							ID: id,
							Price: 100,
						}, nil

					case productID2:
						return&product.Product{
							ID: id,
							Price: 200,
						}, nil
					}

					return nil, product.ErrProductNotFound
				},
			},

			checkResult: func(t *testing.T, cart *domain.Cart) {
				require.NotNil(t, cart)

				require.Len(t, cart.Items, 2)

				require.Equal(t, 2, cart.Items[0].Quantity)
				require.Equal(t, 5, cart.Items[1].Quantity)

				require.Equal(t, userID, cart.Items[0].UserID)
				require.Equal(t, userID, cart.Items[1].UserID)

				require.EqualValues(t, 1200, cart.TotalPrice)
			},
		},
		{
			name: "invalid user id",

			userID: uuid.Nil,

			repo: &domain.MockOrderRepository{},

			product: &domain.MockProductClient{},

			wantErr: ErrInvalidUserID,
		},
		{
			name: "repository error",

			userID: userID,

			repo: &domain.MockOrderRepository{
				GetCartFn: func(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error) {
					return nil, domain.ErrCartEmpty
				},
			},

			product: &domain.MockProductClient{},

			wantErr: domain.ErrCartEmpty,
		},
		{
			name: "product service error",

			userID: userID,

			repo: &domain.MockOrderRepository{
				GetCartFn: func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{
						{
							ID: uuid.New(),
							UserID: id,
							ProductID: productID1,
							Quantity: 1,
						},
					}, nil
				},
			},

			product: &domain.MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*product.Product, error) {
					return nil, product.ErrProductNotFound
				},
			},

			wantErr: product.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, tt.product)

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
			require.Equal(t, 2, tt.product.GetProductFnCalls)
		})
	}
}

func TestService_RemoveFromCart(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	productId := uuid.New()

	tests := []struct {
		name string
		userID uuid.UUID
		productID uuid.UUID
		mock *domain.MockOrderRepository
		wantErr error
	}{
		{
			name: "success",

			userID: userID,
			productID: productId,

			mock: &domain.MockOrderRepository{
				RemoveFromCartFn: func(ctx context.Context, userID, productID uuid.UUID) error {
					return nil
				},
			},
		},
		{
			name: "invalid user id",

			userID: uuid.Nil,
			productID: productId,

			mock: &domain.MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "invalid product id",

			userID: userID,
			productID: uuid.Nil,

			mock: &domain.MockOrderRepository{},
			wantErr: ErrInvalidProductID,
		},
		{
			name: "repository error",

			userID: userID,
			productID: productId,

			mock: &domain.MockOrderRepository{
				RemoveFromCartFn: func(ctx context.Context, userID, productID uuid.UUID) error {
					return domain.ErrProductNotInCart
				},
			},

			wantErr: domain.ErrProductNotInCart,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock, &domain.MockProductClient{})

			err := svc.RemoveFromCart(context.Background(), tt.userID, tt.productID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				if tt.userID != uuid.Nil &&
					tt.productID != uuid.Nil {

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
		mock *domain.MockOrderRepository
		wantErr error
	}{
		{
			name: "success",
			userID: userID,

			mock: &domain.MockOrderRepository{
				ClearCartFn: func(ctx context.Context, userID uuid.UUID) error {
					return nil
				},
			},
		},
		{
			name: "invalid user id",
			userID: uuid.Nil,
			mock: &domain.MockOrderRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "repository error",
			userID: userID,

			mock: &domain.MockOrderRepository{
				ClearCartFn: func(ctx context.Context, userID uuid.UUID) error {
					return domain.ErrCartEmpty
				},
			},

			wantErr: domain.ErrCartEmpty,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.mock, &domain.MockProductClient{})

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
		name string
		orderID uuid.UUID
		mock *domain.MockOrderRepository
		checkResult func(t *testing.T, order *domain.Order)
		wantErr error
	}{
		{
			name: "success",
			orderID: orderID,
			mock: &domain.MockOrderRepository{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID: id,
						UserID: userID,
						Status: domain.StatusCreated,
						Total: 1500,
						CreatedAt: time.Now(),
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
			name: "invalid id",
			orderID: uuid.Nil,
			
			mock: &domain.MockOrderRepository{},

			wantErr: ErrInvalidID,
		},
		{
			name: "repository error",
			orderID: orderID,

			mock: &domain.MockOrderRepository{
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

			svc := newTestService(tt.mock, &domain.MockProductClient{})

			o, err := svc.GetOrderByID(context.Background(), tt.orderID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, o)

				if tt.orderID != uuid.Nil {
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

	tests := []struct {
		name        string
		orderID     uuid.UUID
		mock        *domain.MockOrderRepository
		checkResult func(t *testing.T, items []domain.OrderItem)
		wantErr     error
	}{
		{
			name:    "success",
			orderID: orderID,

			mock: &domain.MockOrderRepository{
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
			name:    "invalid order id",
			orderID: uuid.Nil,

			mock: &domain.MockOrderRepository{},

			wantErr: ErrInvalidOrderID,
		},
		{
			name:    "repository error",
			orderID: orderID,

			mock: &domain.MockOrderRepository{
				GetOrderItemsFn: func(ctx context.Context, id uuid.UUID) ([]domain.OrderItem, error) {
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

			svc := newTestService(tt.mock, &domain.MockProductClient{})

			items, err := svc.GetOrderItems(
				context.Background(),
				tt.orderID,
			)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, items)

				if tt.orderID != uuid.Nil {
					require.Equal(t, 1, tt.mock.GetOrderItemsFnCalls)
				}

				return
			}

			require.NoError(t, err)

			tt.checkResult(t, items)

			require.Equal(t, 1, tt.mock.GetOrderItemsFnCalls)
		})
	}
}

func TestService_CreateOrder(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	productID := uuid.New()

	successRepo := &domain.MockOrderRepository{}

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

	successProduct := &domain.MockProductClient{
		GetProductFn: func(ctx context.Context, id uuid.UUID) (*product.Product, error) {
			return &product.Product{
				ID:           id,
				Name:         "iPhone",
				Price:        120000,
				DeliveryDays: 3,
			}, nil
		},
	}

	tests := []struct {
		name        string
		userID      uuid.UUID
		repo        *domain.MockOrderRepository
		product     *domain.MockProductClient
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
			repo:    &domain.MockOrderRepository{},
			product: &domain.MockProductClient{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:   "empty cart",
			userID: userID,
			repo: &domain.MockOrderRepository{
				GetCartFn: func(ctx context.Context, id uuid.UUID) ([]domain.CartItem, error) {
					return []domain.CartItem{}, nil
				},
			},
			product: &domain.MockProductClient{},
			wantErr: domain.ErrCartEmpty,
		},
		{
			name:   "product not found",
			userID: userID,
			repo: &domain.MockOrderRepository{
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
			product: &domain.MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*product.Product, error) {
					return nil, product.ErrProductNotFound
				},
			},
			wantErr: ErrProductNotFound,
		},
		{
			name:   "transaction error",
			userID: userID,
			product: &domain.MockProductClient{
				GetProductFn: func(ctx context.Context, id uuid.UUID) (*product.Product, error) {
					return &product.Product{
						ID:           id,
						Name:         "iPhone",
						Price:        100,
						DeliveryDays: 2,
					}, nil
				},
			},
			repo: func() *domain.MockOrderRepository {
				repo := &domain.MockOrderRepository{}
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
		tt := tt

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