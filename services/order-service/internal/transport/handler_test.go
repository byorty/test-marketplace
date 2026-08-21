package transport

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/byorty/test-marketplace/services/common/auth"
	rbac "github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/order-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/order-service/internal/generated/openapi"
	"github.com/byorty/test-marketplace/services/order-service/internal/mocks"
	"github.com/byorty/test-marketplace/services/order-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestLogger() *zap.Logger {
	return zap.NewNop()
}

type mockAuthorizer struct {
	allow bool
	err   error
}

func (m *mockAuthorizer) Authorize(role string, resource rbac.Resource, action rbac.Action) error {
	if m.err != nil {
		return m.err
	}
	if !m.allow {
		return rbac.ErrAccessDenied
	}
	return nil
}

func TestHandler_AddToCart(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	productID := uuid.New()
	srvcError := errors.New("service error")

	tests := []struct {
		name        string
		req         api.AddToCartRequestObject
		ctx         context.Context
		mock        *mocks.MockOrderService
		checkResult func(t *testing.T, resp api.AddToCartResponseObject)
	}{
		{
			name: "success",
			req: api.AddToCartRequestObject{
				Body: &api.AddToCartJSONRequestBody{
					ProductId: productID,
					Quantity:  2,
				},
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				AddToCartFunc: func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
					require.Equal(t, userID, item.UserID)
					require.Equal(t, productID, item.ProductID)
					require.Equal(t, 2, item.Quantity)
					return nil
				},
			},
			checkResult: func(t *testing.T, resp api.AddToCartResponseObject) {
				_, ok := resp.(api.AddToCart201Response)
				require.True(t, ok)
			},
		},
		{
			name: "unauthorized - no claims",
			req: api.AddToCartRequestObject{
				Body: &api.AddToCartJSONRequestBody{
					ProductId: productID,
					Quantity:  1,
				},
			},
			ctx: context.Background(),
			mock: &mocks.MockOrderService{},
			checkResult: func(t *testing.T, resp api.AddToCartResponseObject) {
				require.IsType(t, api.AddToCart401JSONResponse{}, resp)
			},
		},
		{
			name: "validation error - empty product id",
			req: api.AddToCartRequestObject{
				Body: &api.AddToCartJSONRequestBody{
					ProductId: uuid.Nil,
					Quantity:  1,
				},
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				AddToCartFunc: func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
					return service.ErrInvalidInput
				},
			},
			checkResult: func(t *testing.T, resp api.AddToCartResponseObject) {
				require.IsType(t, api.AddToCart400JSONResponse{}, resp)
			},
		},
		{
			name: "not found - product doesn't exist",
			req: api.AddToCartRequestObject{
				Body: &api.AddToCartJSONRequestBody{
					ProductId: productID,
					Quantity:  1,
				},
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				AddToCartFunc: func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
					return service.ErrProductNotFound
				},
			},
			checkResult: func(t *testing.T, resp api.AddToCartResponseObject) {
				require.IsType(t, api.AddToCart404JSONResponse{}, resp)
			},
		},
		{
			name: "internal error",
			req: api.AddToCartRequestObject{
				Body: &api.AddToCartJSONRequestBody{
					ProductId: productID,
					Quantity:  1,
				},
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				AddToCartFunc: func(ctx context.Context, userID uuid.UUID, item *domain.CartItem) error {
					return srvcError
				},
			},
			checkResult: func(t *testing.T, resp api.AddToCartResponseObject) {
				require.IsType(t, api.AddToCart500JSONResponse{}, resp)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &OrderHandler{
				service: tt.mock,
				log:     newTestLogger(),
			}

			resp, err := handler.AddToCart(tt.ctx, tt.req)

			require.NoError(t, err)
			tt.checkResult(t, resp)

			if tt.name == "unauthorized - no claims" {
				require.Zero(t, tt.mock.AddToCartCalls)
			} else {
				require.Equal(t, 1, tt.mock.AddToCartCalls)
			}

			require.Zero(t, tt.mock.GetCartCalls)
			require.Zero(t, tt.mock.RemoveFromCartCalls)
			require.Zero(t, tt.mock.GetOrderByIDCalls)
			require.Zero(t, tt.mock.CreateOrderCalls)
		})
	}
}

func TestHandler_CreateOrder(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	orderID := uuid.New()
	now := time.Now()
	srvcError := errors.New("service error")

	tests := []struct {
		name        string
		req         api.CreateOrderRequestObject
		ctx         context.Context
		mock        *mocks.MockOrderService
		checkResult func(t *testing.T, resp api.CreateOrderResponseObject)
	}{
		{
			name: "success",
			req:  api.CreateOrderRequestObject{},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				CreateOrderFunc: func(ctx context.Context, uid uuid.UUID) (*domain.Order, error) {
					require.Equal(t, userID, uid)
					return &domain.Order{
						ID:           orderID,
						UserID:       userID,
						Status:       domain.Status("pending"),
						TotalPrice:        150000,
						CreatedAt:    now,
						DeliveryDate: now.Add(5 * 24 * time.Hour),
					}, nil
				},
			},
			checkResult: func(t *testing.T, resp api.CreateOrderResponseObject) {
				response, ok := resp.(api.CreateOrder201JSONResponse)
				require.True(t, ok)
				require.Equal(t, orderID, response.Id)
				require.Equal(t, api.OrderStatus("pending"), response.Status)
			},
		},
		{
			name: "unauthorized - no claims",
			req:  api.CreateOrderRequestObject{},
			ctx:  context.Background(),
			mock: &mocks.MockOrderService{},
			checkResult: func(t *testing.T, resp api.CreateOrderResponseObject) {
				require.IsType(t, api.CreateOrder401JSONResponse{}, resp)
			},
		},
		{
			name: "validation error - invalid user id",
			req:  api.CreateOrderRequestObject{},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				CreateOrderFunc: func(ctx context.Context, uid uuid.UUID) (*domain.Order, error) {
					return nil, service.ErrInvalidUserID
				},
			},
			checkResult: func(t *testing.T, resp api.CreateOrderResponseObject) {
				require.IsType(t, api.CreateOrder400JSONResponse{}, resp)
			},
		},
		{
			name: "cart empty",
			req:  api.CreateOrderRequestObject{},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				CreateOrderFunc: func(ctx context.Context, uid uuid.UUID) (*domain.Order, error) {
					return nil, domain.ErrCartEmpty
				},
			},
			checkResult: func(t *testing.T, resp api.CreateOrderResponseObject) {
				require.IsType(t, api.CreateOrder400JSONResponse{}, resp)
			},
		},
		{
			name: "internal error",
			req:  api.CreateOrderRequestObject{},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				CreateOrderFunc: func(ctx context.Context, uid uuid.UUID) (*domain.Order, error) {
					return nil, srvcError
				},
			},
			checkResult: func(t *testing.T, resp api.CreateOrderResponseObject) {
				require.IsType(t, api.CreateOrder500JSONResponse{}, resp)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &OrderHandler{
				service: tt.mock,
				log:     newTestLogger(),
			}

			resp, err := handler.CreateOrder(tt.ctx, tt.req)

			require.NoError(t, err)
			tt.checkResult(t, resp)

			if tt.name == "unauthorized - no claims" {
				require.Zero(t, tt.mock.CreateOrderCalls)
			} else {
				require.Equal(t, 1, tt.mock.CreateOrderCalls)
			}

			require.Zero(t, tt.mock.AddToCartCalls)
			require.Zero(t, tt.mock.GetCartCalls)
			require.Zero(t, tt.mock.RemoveFromCartCalls)
			require.Zero(t, tt.mock.GetOrderByIDCalls)
		})
	}
}

func TestHandler_GetCart(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	productID1 := uuid.New()
	productID2 := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()
	srvcError := errors.New("service error")

	tests := []struct {
		name        string
		req         api.GetCartRequestObject
		ctx         context.Context
		mock        *mocks.MockOrderService
		checkResult func(t *testing.T, resp api.GetCartResponseObject)
	}{
		{
			name: "success",
			req:  api.GetCartRequestObject{},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				GetCartFunc: func(ctx context.Context, uid uuid.UUID) (*domain.Cart, error) {
					require.Equal(t, userID, uid)
					return &domain.Cart{
						Items: []domain.CartItem{
							{
								ID:        itemID1,
								UserID:    userID,
								ProductID: productID1,
								Quantity:  2,
							},
							{
								ID:        itemID2,
								UserID:    userID,
								ProductID: productID2,
								Quantity:  1,
							},
						},
						TotalPrice: 150000,
					}, nil
				},
			},
			checkResult: func(t *testing.T, resp api.GetCartResponseObject) {
				response, ok := resp.(api.GetCart200JSONResponse)
				require.True(t, ok)
				require.Len(t, response.Items, 2)
				require.Equal(t, productID1, response.Items[0].ProductId)
				require.Equal(t, int32(2), response.Items[0].Quantity)
				require.Equal(t, productID2, response.Items[1].ProductId)
				require.Equal(t, int32(1), response.Items[1].Quantity)
				require.Equal(t, int64(150000), response.TotalPrice)
			},
		},
		{
			name: "success - empty cart",
			req:  api.GetCartRequestObject{},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				GetCartFunc: func(ctx context.Context, uid uuid.UUID) (*domain.Cart, error) {
					return &domain.Cart{
						Items:      []domain.CartItem{},
						TotalPrice: 0,
					}, nil
				},
			},
			checkResult: func(t *testing.T, resp api.GetCartResponseObject) {
				response, ok := resp.(api.GetCart200JSONResponse)
				require.True(t, ok)
				require.Len(t, response.Items, 0)
				require.Equal(t, int64(0), response.TotalPrice)
			},
		},
		{
			name: "unauthorized - no claims",
			req:  api.GetCartRequestObject{},
			ctx:  context.Background(),
			mock: &mocks.MockOrderService{},
			checkResult: func(t *testing.T, resp api.GetCartResponseObject) {
				require.IsType(t, api.GetCart401JSONResponse{}, resp)
			},
		},
		{
			name: "internal error",
			req:  api.GetCartRequestObject{},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: userID,
			}),
			mock: &mocks.MockOrderService{
				GetCartFunc: func(ctx context.Context, uid uuid.UUID) (*domain.Cart, error) {
					return nil, srvcError
				},
			},
			checkResult: func(t *testing.T, resp api.GetCartResponseObject) {
				require.IsType(t, api.GetCart500JSONResponse{}, resp)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &OrderHandler{
				service: tt.mock,
				log:     newTestLogger(),
			}

			resp, err := handler.GetCart(tt.ctx, tt.req)

			require.NoError(t, err)
			tt.checkResult(t, resp)

			if tt.name == "unauthorized - no claims" {
				require.Zero(t, tt.mock.GetCartCalls)
			} else {
				require.Equal(t, 1, tt.mock.GetCartCalls)
			}

			require.Zero(t, tt.mock.AddToCartCalls)
			require.Zero(t, tt.mock.RemoveFromCartCalls)
			require.Zero(t, tt.mock.GetOrderByIDCalls)
			require.Zero(t, tt.mock.CreateOrderCalls)
		})
	}
}



func TestHandler_GetOrderByID(t *testing.T) {
	t.Parallel()
	productID1 := uuid.New()
	productID2 := uuid.New()
	now := time.Now()
	deliveryDate := now.Add(5 * 24 * time.Hour)
	srvcError := errors.New("service error")
	expectedOrderID := uuid.New()
	expectedUserID := uuid.New()

	allowAll := &mockAuthorizer{allow: true}
	denyAll := &mockAuthorizer{allow: false}

	tests := []struct {
		name        string
		req         api.GetOrderByIDRequestObject
		ctx         context.Context
		mock        *mocks.MockOrderService
		authorizer  domain.Authorizer
		checkResult func(t *testing.T, resp api.GetOrderByIDResponseObject)
	}{
		{
			name: "success - owner",
			req: api.GetOrderByIDRequestObject{
				Id: expectedOrderID,
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: expectedUserID,
				Role:   "user",
			}),
			mock: &mocks.MockOrderService{
				GetOrderByIDFunc: func(ctx context.Context, userID, orderID uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:           orderID,
						UserID:       expectedUserID,
						Status:       domain.Status("pending"),
						TotalPrice:        150000,
						CreatedAt:    now,
						DeliveryDate: deliveryDate,
						Items: []domain.OrderItem{
							{
								ID:           uuid.New(),
								OrderID:      orderID,
								ProductID:    productID1,
								ProductPrice: 50000,
								Quantity:     1,
							},
							{
								ID:           uuid.New(),
								OrderID:      orderID,
								ProductID:    productID2,
								ProductPrice: 100000,
								Quantity:     2,
							},
						},
					}, nil
				},
			},
			authorizer: allowAll,
			checkResult: func(t *testing.T, resp api.GetOrderByIDResponseObject) {
				response, ok := resp.(api.GetOrderByID200JSONResponse)
				require.True(t, ok)
				require.Equal(t, expectedOrderID, response.Id)
				require.Equal(t, api.OrderStatus("pending"), response.Status)
				require.Equal(t, int64(150000), response.TotalPrice)
				require.Len(t, response.Items, 2)
				require.Equal(t, productID1, response.Items[0].ProductId)
				require.Equal(t, int64(50000), response.Items[0].ProductPrice)
				require.Equal(t, int32(1), response.Items[0].Quantity)
				require.Equal(t, productID2, response.Items[1].ProductId)
				require.Equal(t, int64(100000), response.Items[1].ProductPrice)
				require.Equal(t, int32(2), response.Items[1].Quantity)
			},
		},
		{
			name: "success - employee",
			req: api.GetOrderByIDRequestObject{
				Id: expectedOrderID,
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: uuid.New(),
				Role:   "employee",
			}),
			mock: &mocks.MockOrderService{
				GetOrderByIDFunc: func(ctx context.Context, userID, orderID uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:           orderID,
						UserID:       expectedUserID,
						Status:       domain.Status("pending"),
						TotalPrice:        150000,
						CreatedAt:    now,
						DeliveryDate: deliveryDate,
						Items:        []domain.OrderItem{},
					}, nil
				},
			},
			authorizer: allowAll,
			checkResult: func(t *testing.T, resp api.GetOrderByIDResponseObject) {
				_, ok := resp.(api.GetOrderByID200JSONResponse)
				require.True(t, ok)
			},
		},
		{
			name: "unauthorized - no claims",
			req: api.GetOrderByIDRequestObject{
				Id: expectedOrderID,
			},
			ctx:        context.Background(),
			mock:       &mocks.MockOrderService{},
			authorizer: allowAll,
			checkResult: func(t *testing.T, resp api.GetOrderByIDResponseObject) {
				require.IsType(t, api.GetOrderByID401JSONResponse{}, resp)
			},
		},
		{
			name: "forbidden - rbac access denied",
			req: api.GetOrderByIDRequestObject{
				Id: expectedOrderID,
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: expectedUserID,
				Role:   "guest",
			}),
			mock:       &mocks.MockOrderService{},
			authorizer: denyAll,
			checkResult: func(t *testing.T, resp api.GetOrderByIDResponseObject) {
				require.IsType(t, api.GetOrderByID403JSONResponse{}, resp)
			},
		},
		{
			name: "forbidden - not owner and not employee",
			req: api.GetOrderByIDRequestObject{
				Id: expectedOrderID,
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: uuid.New(),
				Role:   "user",
			}),
			mock: &mocks.MockOrderService{
				GetOrderByIDFunc: func(ctx context.Context, userID, orderID uuid.UUID) (*domain.Order, error) {
					return &domain.Order{
						ID:     orderID,
						UserID: expectedUserID,
						Status: domain.Status("pending"),
						TotalPrice:  150000,
					}, nil
				},
			},
			authorizer: allowAll,
			checkResult: func(t *testing.T, resp api.GetOrderByIDResponseObject) {
				require.IsType(t, api.GetOrderByID403JSONResponse{}, resp)
			},
		},
		{
			name: "order not found",
			req: api.GetOrderByIDRequestObject{
				Id: expectedOrderID,
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: expectedUserID,
				Role:   "user",
			}),
			mock: &mocks.MockOrderService{
				GetOrderByIDFunc: func(ctx context.Context, userID, orderID uuid.UUID) (*domain.Order, error) {
					return nil, domain.ErrOrderNotFound
				},
			},
			authorizer: allowAll,
			checkResult: func(t *testing.T, resp api.GetOrderByIDResponseObject) {
				require.IsType(t, api.GetOrderByID404JSONResponse{}, resp)
			},
		},
		{
			name: "internal error",
			req: api.GetOrderByIDRequestObject{
				Id: expectedOrderID,
			},
			ctx: auth.ContextWithClaims(context.Background(), &auth.Claims{
				UserID: expectedUserID,
				Role:   "user",
			}),
			mock: &mocks.MockOrderService{
				GetOrderByIDFunc: func(ctx context.Context, userID, orderID uuid.UUID) (*domain.Order, error) {
					return nil, srvcError
				},
			},
			authorizer: allowAll,
			checkResult: func(t *testing.T, resp api.GetOrderByIDResponseObject) {
				require.IsType(t, api.GetOrderByID500JSONResponse{}, resp)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &OrderHandler{
				service:    tt.mock,
				log:        newTestLogger(),
				authorizer: tt.authorizer,
			}

			resp, err := handler.GetOrderByID(tt.ctx, tt.req)

			require.NoError(t, err)
			tt.checkResult(t, resp)

			noServiceCall := tt.name == "unauthorized - no claims" || tt.name == "forbidden - rbac access denied"
			if noServiceCall {
				require.Zero(t, tt.mock.GetOrderByIDCalls)
			} else {
				require.Equal(t, 1, tt.mock.GetOrderByIDCalls)
			}

			require.Zero(t, tt.mock.AddToCartCalls)
			require.Zero(t, tt.mock.GetCartCalls)
			require.Zero(t, tt.mock.RemoveFromCartCalls)
			require.Zero(t, tt.mock.CreateOrderCalls)
		})
	}
}

func TestHandler_RemoveFromCart(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	cartItemID := uuid.New()
	serviceError := errors.New("service error")

	tests := []struct {
		name        string
		req         api.RemoveFromCartRequestObject
		ctx         context.Context
		mock        *mocks.MockOrderService
		checkResult func(t *testing.T, resp api.RemoveFromCartResponseObject)
	}{
		{
			name: "success",
			req: api.RemoveFromCartRequestObject{
				Id: cartItemID,
			},
			ctx: auth.ContextWithClaims(
				context.Background(),
				&auth.Claims{
					UserID: userID,
				},
			),
			mock: &mocks.MockOrderService{
				RemoveFromCartFunc: func(
					ctx context.Context,
					uid uuid.UUID,
					id uuid.UUID,
				) error {
					require.Equal(t, userID, uid)
					require.Equal(t, id, cartItemID)

					return nil
				},
			},
			checkResult: func(
				t *testing.T,
				resp api.RemoveFromCartResponseObject,
			) {
				_, ok := resp.(api.RemoveFromCart204Response)
				require.True(t, ok)
			},
		},
		{
			name: "unauthorized - no claims",
			req: api.RemoveFromCartRequestObject{
				Id: cartItemID,
			},
			ctx:  context.Background(),
			mock: &mocks.MockOrderService{},
			checkResult: func(
				t *testing.T,
				resp api.RemoveFromCartResponseObject,
			) {
				require.IsType(
					t,
					api.RemoveFromCart401JSONResponse{},
					resp,
				)
			},
		},
		{
			name: "not found - cart item not found",
			req: api.RemoveFromCartRequestObject{
				Id: cartItemID,
			},
			ctx: auth.ContextWithClaims(
				context.Background(),
				&auth.Claims{
					UserID: userID,
				},
			),
			mock: &mocks.MockOrderService{
				RemoveFromCartFunc: func(
					ctx context.Context,
					uid uuid.UUID,
					cartItemID uuid.UUID,
				) error {
					return domain.ErrCartItemNotFound
				},
			},
			checkResult: func(
				t *testing.T,
				resp api.RemoveFromCartResponseObject,
			) {
				require.IsType(
					t,
					api.RemoveFromCart404JSONResponse{},
					resp,
				)
			},
		},
		{
			name: "internal error",
			req: api.RemoveFromCartRequestObject{
				Id: cartItemID,
			},
			ctx: auth.ContextWithClaims(
				context.Background(),
				&auth.Claims{
					UserID: userID,
				},
			),
			mock: &mocks.MockOrderService{
				RemoveFromCartFunc: func(
					ctx context.Context,
					uid uuid.UUID,
					cartItemID uuid.UUID,
				) error {
					return serviceError
				},
			},
			checkResult: func(
				t *testing.T,
				resp api.RemoveFromCartResponseObject,
			) {
				require.IsType(
					t,
					api.RemoveFromCart500JSONResponse{},
					resp,
				)
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &OrderHandler{
				service: tt.mock,
				log:     newTestLogger(),
			}

			resp, err := handler.RemoveFromCart(tt.ctx, tt.req)

			require.NoError(t, err)
			tt.checkResult(t, resp)

			if tt.name == "unauthorized - no claims" {
				require.Zero(t, tt.mock.RemoveFromCartCalls)
			} else {
				require.Equal(t, 1, tt.mock.RemoveFromCartCalls)
			}

			require.Zero(t, tt.mock.AddToCartCalls)
			require.Zero(t, tt.mock.GetCartCalls)
			require.Zero(t, tt.mock.GetOrderByIDCalls)
			require.Zero(t, tt.mock.CreateOrderCalls)
		})
	}
}