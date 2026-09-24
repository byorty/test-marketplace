package transport

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/byorty/test-marketplace/services/common/auth"
	orderclient "github.com/byorty/test-marketplace/services/common/client/order"
	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/delivery-service/internal/generated/openapi"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/mocks"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/qr"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestHandler(repo *mocks.MockDeliveryRepository, orderClient *mocks.MockOrderClient, auth *mocks.MockAuthorizer) *DeliveryHandler {
	qrGen, err := qr.NewGenerator("test-secret-key-that-is-32-bytes-long!!", 24*time.Hour)
	if err != nil {
		panic(err)
	}
	qrVer, err := qr.NewVerifier("test-secret-key-that-is-32-bytes-long!!")
	if err != nil {
		panic(err)
	}
	svc := service.New(
		zap.NewNop(),
		repo,
		orderClient,
		qrGen,
		qrVer,
		validator.New(),
		auth,
		&service.RealClock{},
		service.NewLoggingNotifier(zap.NewNop()),
	)
	return &DeliveryHandler{
		service:    svc,
		log:        zap.NewNop(),
		authorizer: auth,
	}
}

func allowAllAuthorizer() *mocks.MockAuthorizer {
	return &mocks.MockAuthorizer{
		AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
			return nil
		},
	}
}

func TestHandler_CreateDelivery(t *testing.T) {
	t.Parallel()

	orderID := uuid.New()
	userID := uuid.New()
	futureDate := time.Now().Add(48 * time.Hour)

	tests := []struct {
		name       string
		role       string
		body       api.DeliveryCreateRequest
		repo       *mocks.MockDeliveryRepository
		client     *mocks.MockOrderClient
		auth       *mocks.MockAuthorizer
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			role: "employee",
			body: api.DeliveryCreateRequest{
				OrderId:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByOrderIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
				CreateFn: func(ctx context.Context, d *domain.Delivery) error {
					return nil
				},
			},
			client: &mocks.MockOrderClient{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*orderclient.OrderResponse, error) {
					return &orderclient.OrderResponse{ID: id, UserID: userID, Status: "PAID"}, nil
				},
			},
			auth:       allowAllAuthorizer(),
			wantStatus: 201,
		},
		{
			name: "customer forbidden",
			role: "customer",
			body: api.DeliveryCreateRequest{
				OrderId:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
			},
			repo:   &mocks.MockDeliveryRepository{},
			client: &mocks.MockOrderClient{},
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if role == "customer" && action == rbac.ActionCreate {
						return rbac.ErrAccessDenied
					}
					return nil
				},
			},
			wantStatus: 403,
			wantCode:   "forbidden",
		},
		{
			name: "invalid pickup address",
			role: "employee",
			body: api.DeliveryCreateRequest{
				OrderId:               orderID,
				PickupAddress:         "",
				EstimatedDeliveryDate: futureDate,
			},
			repo:       &mocks.MockDeliveryRepository{},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 400,
			wantCode:   "validation_error",
		},
		{
			name: "delivery already exists",
			role: "employee",
			body: api.DeliveryCreateRequest{
				OrderId:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByOrderIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: uuid.New(), Status: domain.DeliveryStatusInTransit}, nil
				},
			},
			client: &mocks.MockOrderClient{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*orderclient.OrderResponse, error) {
					return &orderclient.OrderResponse{ID: id, UserID: userID, Status: "PAID"}, nil
				},
			},
			auth:       allowAllAuthorizer(),
			wantStatus: 409,
			wantCode:   "delivery_already_exists",
		},
		{
			name: "order not paid",
			role: "employee",
			body: api.DeliveryCreateRequest{
				OrderId:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
			},
			repo: &mocks.MockDeliveryRepository{},
			client: &mocks.MockOrderClient{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*orderclient.OrderResponse, error) {
					return &orderclient.OrderResponse{ID: id, UserID: userID, Status: "CREATED"}, nil
				},
			},
			auth:       allowAllAuthorizer(),
			wantStatus: 400,
			wantCode:   "order_not_paid",
		},
		{
			name: "order not found",
			role: "employee",
			body: api.DeliveryCreateRequest{
				OrderId:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
			},
			repo: &mocks.MockDeliveryRepository{},
			client: &mocks.MockOrderClient{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*orderclient.OrderResponse, error) {
					return nil, orderclient.ErrOrderNotFound
				},
			},
			auth:       allowAllAuthorizer(),
			wantStatus: 500,
			wantCode:   "internal_error",
		},
		{
			name: "database failure",
			role: "employee",
			body: api.DeliveryCreateRequest{
				OrderId:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByOrderIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
				CreateFn: func(ctx context.Context, d *domain.Delivery) error {
					return errors.New("db error")
				},
			},
			client: &mocks.MockOrderClient{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*orderclient.OrderResponse, error) {
					return &orderclient.OrderResponse{ID: id, UserID: userID, Status: "PAID"}, nil
				},
			},
			auth:       allowAllAuthorizer(),
			wantStatus: 500,
			wantCode:   "internal_error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestHandler(tt.repo, tt.client, tt.auth)
			req := api.CreateDeliveryRequestObject{Body: &tt.body}

			claims := &auth.Claims{UserID: userID, Role: tt.role}
			ctx := auth.ContextWithClaims(context.Background(), claims)

			resp, err := handler.CreateDelivery(ctx, req)
			require.NoError(t, err)

			switch tt.wantStatus {
			case 201:
				r, ok := resp.(api.CreateDelivery201JSONResponse)
				require.True(t, ok, "expected 201 response, got %T", resp)
				require.NotEqual(t, uuid.Nil, r.Id)
				require.Equal(t, domain.DeliveryStatusPending, domain.DeliveryStatus(r.Status))
			case 400:
				r, ok := resp.(api.CreateDelivery400JSONResponse)
				require.True(t, ok, "expected 400 response, got %T", resp)
				if tt.wantCode != "" {
					require.Equal(t, tt.wantCode, string(r.Code))
				}
			case 403:
				r, ok := resp.(api.CreateDelivery403JSONResponse)
				require.True(t, ok, "expected 403 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 409:
				r, ok := resp.(api.CreateDelivery409JSONResponse)
				require.True(t, ok, "expected 409 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 500:
				r, ok := resp.(api.CreateDelivery500JSONResponse)
				require.True(t, ok, "expected 500 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			}
		})
	}
}

func TestHandler_GetDelivery(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	customerID := uuid.New()
	otherCustomerID := uuid.New()
	orderID := uuid.New()

	tests := []struct {
		name       string
		role       string
		userID     uuid.UUID
		repo       *mocks.MockDeliveryRepository
		client     *mocks.MockOrderClient
		auth       *mocks.MockAuthorizer
		wantStatus int
		wantCode   string
	}{
		{
			name:   "customer views own delivery",
			role:   "customer",
			userID: customerID,
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if role == "customer" && action == rbac.ActionView {
						return nil
					}
					return rbac.ErrAccessDenied
				},
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:                    deliveryID,
						UserID:                customerID,
						OrderID:               orderID,
						Status:                domain.DeliveryStatusInTransit,
						PickupAddress:         "123 Main St",
						EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
						IsLate:                false,
						CreatedAt:             time.Now(),
						UpdatedAt:             time.Now(),
					}, nil
				},
				ListReschedulesFn: func(ctx context.Context, dID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			wantStatus: 200,
		},
		{
			name:   "employee views any delivery",
			role:   "employee",
			userID: uuid.New(),
			auth:   allowAllAuthorizer(),
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:                    deliveryID,
						UserID:                customerID,
						OrderID:               orderID,
						Status:                domain.DeliveryStatusPending,
						PickupAddress:         "456 Oak Ave",
						EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
						IsLate:                false,
						CreatedAt:             time.Now(),
						UpdatedAt:             time.Now(),
					}, nil
				},
				ListReschedulesFn: func(ctx context.Context, dID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			wantStatus: 200,
		},
		{
			name:   "IDOR: customer B cannot view customer A delivery",
			role:   "customer",
			userID: otherCustomerID,
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					return rbac.ErrAccessDenied
				},
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     deliveryID,
						UserID: customerID,
						Status: domain.DeliveryStatusInTransit,
					}, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			wantStatus: 403,
			wantCode:   "forbidden",
		},
		{
			name:   "delivery not found",
			role:   "customer",
			userID: customerID,
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					return nil
				},
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
			},
			client:     &mocks.MockOrderClient{},
			wantStatus: 404,
			wantCode:   "delivery_not_found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestHandler(tt.repo, tt.client, tt.auth)
			req := api.GetDeliveryRequestObject{Id: deliveryID}

			claims := &auth.Claims{UserID: tt.userID, Role: tt.role}
			ctx := auth.ContextWithClaims(context.Background(), claims)

			resp, err := handler.GetDelivery(ctx, req)
			require.NoError(t, err)

			switch tt.wantStatus {
			case 200:
				r, ok := resp.(api.GetDelivery200JSONResponse)
				require.True(t, ok, "expected 200 response, got %T", resp)
				require.Equal(t, deliveryID, r.Id)
				require.Equal(t, customerID, r.UserId)
				require.Equal(t, orderID, r.OrderId)
			case 403:
				r, ok := resp.(api.GetDelivery403JSONResponse)
				require.True(t, ok, "expected 403 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 404:
				r, ok := resp.(api.GetDelivery404JSONResponse)
				require.True(t, ok, "expected 404 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			}
		})
	}
}

func TestHandler_RescheduleDelivery(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	futureDate := time.Now().Add(72 * time.Hour)

	tests := []struct {
		name       string
		role       string
		body       api.DeliveryRescheduleRequest
		repo       *mocks.MockDeliveryRepository
		client     *mocks.MockOrderClient
		auth       *mocks.MockAuthorizer
		wantStatus int
		wantCode   string
	}{
		{
			name: "success - employee reschedule",
			role: "employee",
			body: api.DeliveryRescheduleRequest{
				NewDate: futureDate,
				Reason:  strPtr("weather delay"),
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:                    deliveryID,
						UserID:                uuid.New(),
						OrderID:               uuid.New(),
						Status:                domain.DeliveryStatusPending,
						EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
					}, nil
				},
				RescheduleTxFn: func(ctx context.Context, d *domain.Delivery, r *domain.DeliveryReschedule) (*domain.Delivery, error) {
					return d, nil
				},
				ListReschedulesFn: func(ctx context.Context, dID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 200,
		},
		{
			name: "customer forbidden",
			role: "customer",
			body: api.DeliveryRescheduleRequest{
				NewDate: futureDate,
			},
			repo:   &mocks.MockDeliveryRepository{},
			client: &mocks.MockOrderClient{},
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if role == "customer" && action == rbac.ActionReschedule {
						return rbac.ErrAccessDenied
					}
					return nil
				},
			},
			wantStatus: 403,
			wantCode:   "forbidden",
		},
		{
			name: "delivery not found",
			role: "employee",
			body: api.DeliveryRescheduleRequest{
				NewDate: futureDate,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
			},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 404,
			wantCode:   "delivery_not_found",
		},
		{
			name: "terminal status - cannot reschedule",
			role: "employee",
			body: api.DeliveryRescheduleRequest{
				NewDate: futureDate,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     deliveryID,
						Status: domain.DeliveryStatusDelivered,
					}, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 409,
			wantCode:   "cannot_reschedule",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestHandler(tt.repo, tt.client, tt.auth)
			req := api.RescheduleDeliveryRequestObject{
				Id:   deliveryID,
				Body: &tt.body,
			}

			claims := &auth.Claims{UserID: uuid.New(), Role: tt.role}
			ctx := auth.ContextWithClaims(context.Background(), claims)

			resp, err := handler.RescheduleDelivery(ctx, req)
			require.NoError(t, err)

			switch tt.wantStatus {
			case 200:
				r, ok := resp.(api.RescheduleDelivery200JSONResponse)
				require.True(t, ok, "expected 200 response, got %T", resp)
				require.Equal(t, deliveryID, r.Id)
			case 400:
				r, ok := resp.(api.RescheduleDelivery400JSONResponse)
				require.True(t, ok, "expected 400 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 403:
				r, ok := resp.(api.RescheduleDelivery403JSONResponse)
				require.True(t, ok, "expected 403 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 404:
				r, ok := resp.(api.RescheduleDelivery404JSONResponse)
				require.True(t, ok, "expected 404 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 409:
				r, ok := resp.(api.RescheduleDelivery409JSONResponse)
				require.True(t, ok, "expected 409 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestHandler_UpdateDeliveryStatus(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()

	tests := []struct {
		name       string
		role       string
		body       api.DeliveryStatusUpdateRequest
		repo       *mocks.MockDeliveryRepository
		client     *mocks.MockOrderClient
		auth       *mocks.MockAuthorizer
		wantStatus int
		wantCode   string
	}{
		{
			name: "success - pending to in_transit",
			role: "employee",
			body: api.DeliveryStatusUpdateRequest{
				Status: api.INTRANSIT,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:      deliveryID,
						UserID:  uuid.New(),
						OrderID: uuid.New(),
						Status:  domain.DeliveryStatusPending,
					}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 200,
		},
		{
			name: "customer forbidden",
			role: "customer",
			body: api.DeliveryStatusUpdateRequest{
				Status: api.INTRANSIT,
			},
			repo:   &mocks.MockDeliveryRepository{},
			client: &mocks.MockOrderClient{},
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if role == "customer" && action == rbac.ActionUpdateStatus {
						return rbac.ErrAccessDenied
					}
					return nil
				},
			},
			wantStatus: 403,
			wantCode:   "forbidden",
		},
		{
			name: "invalid status value",
			role: "employee",
			body: api.DeliveryStatusUpdateRequest{
				Status: api.DeliveryStatus("INVALID"),
			},
			repo:       &mocks.MockDeliveryRepository{},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 400,
			wantCode:   "validation_error",
		},
		{
			name: "delivery not found",
			role: "employee",
			body: api.DeliveryStatusUpdateRequest{
				Status: api.INTRANSIT,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
			},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 404,
			wantCode:   "delivery_not_found",
		},
		{
			name: "forbidden transition - pending to delivered",
			role: "employee",
			body: api.DeliveryStatusUpdateRequest{
				Status: api.DELIVERED,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 400,
			wantCode:   "invalid_status_transition",
		},
		{
			name: "already delivered",
			role: "employee",
			body: api.DeliveryStatusUpdateRequest{
				Status: api.INTRANSIT,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusDelivered}, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 409,
			wantCode:   "delivery_already_delivered",
		},
		{
			name: "already cancelled",
			role: "employee",
			body: api.DeliveryStatusUpdateRequest{
				Status: api.INTRANSIT,
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusCancelled}, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			auth:       allowAllAuthorizer(),
			wantStatus: 409,
			wantCode:   "delivery_already_cancelled",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestHandler(tt.repo, tt.client, tt.auth)
			req := api.UpdateDeliveryStatusRequestObject{
				Id:   deliveryID,
				Body: &tt.body,
			}

			claims := &auth.Claims{UserID: uuid.New(), Role: tt.role}
			ctx := auth.ContextWithClaims(context.Background(), claims)

			resp, err := handler.UpdateDeliveryStatus(ctx, req)
			require.NoError(t, err)

			switch tt.wantStatus {
			case 200:
				r, ok := resp.(api.UpdateDeliveryStatus200JSONResponse)
				require.True(t, ok, "expected 200 response, got %T", resp)
				require.Equal(t, deliveryID, r.Id)
			case 400:
				r, ok := resp.(api.UpdateDeliveryStatus400JSONResponse)
				require.True(t, ok, "expected 400 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 403:
				r, ok := resp.(api.UpdateDeliveryStatus403JSONResponse)
				require.True(t, ok, "expected 403 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 404:
				r, ok := resp.(api.UpdateDeliveryStatus404JSONResponse)
				require.True(t, ok, "expected 404 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 409:
				r, ok := resp.(api.UpdateDeliveryStatus409JSONResponse)
				require.True(t, ok, "expected 409 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			}
		})
	}
}
func TestHandler_GetDeliveryQR(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	customerID := uuid.New()

	tests := []struct {
		name       string
		role       string
		userID     uuid.UUID
		deliveryID openapi_types.UUID
		auth       *mocks.MockAuthorizer
		repo       *mocks.MockDeliveryRepository
		client     *mocks.MockOrderClient
		wantStatus int
		wantCode   string
	}{
		{
			name:       "customer gets own QR - success",
			role:       "customer",
			userID:     customerID,
			deliveryID: openapi_types.UUID(deliveryID),
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if action == rbac.ActionGetQR {
						return nil
					}
					return rbac.ErrAccessDenied
				},
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:      deliveryID,
						UserID:  customerID,
						Status:  domain.DeliveryStatusReadyForPickup,
						QRNonce: uuid.NullUUID{},
					}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			wantStatus: 200,
		},
		{
			name:       "customer gets another's QR - forbidden",
			role:       "customer",
			userID:     uuid.New(),
			deliveryID: openapi_types.UUID(deliveryID),
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					return rbac.ErrAccessDenied
				},
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     deliveryID,
						UserID: customerID,
						Status: domain.DeliveryStatusReadyForPickup,
					}, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			wantStatus: 403,
			wantCode:   "forbidden",
		},
		{
			name:       "wrong status - conflict",
			role:       "customer",
			userID:     customerID,
			deliveryID: openapi_types.UUID(deliveryID),
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					return nil
				},
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     deliveryID,
						UserID: customerID,
						Status: domain.DeliveryStatusPending,
					}, nil
				},
			},
			client:     &mocks.MockOrderClient{},
			wantStatus: 409,
			wantCode:   "invalid_delivery_status",
		},
		{
			name:       "delivery not found",
			role:       "customer",
			userID:     customerID,
			deliveryID: openapi_types.UUID(deliveryID),
			auth:       allowAllAuthorizer(),
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
			},
			client:     &mocks.MockOrderClient{},
			wantStatus: 404,
			wantCode:   "delivery_not_found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestHandler(tt.repo, tt.client, tt.auth)
			req := api.GetDeliveryQRRequestObject{
				Id: tt.deliveryID,
			}

			claims := &auth.Claims{UserID: tt.userID, Role: tt.role}
			ctx := auth.ContextWithClaims(context.Background(), claims)

			resp, err := handler.GetDeliveryQR(ctx, req)
			require.NoError(t, err)

			switch tt.wantStatus {
			case 200:
				r, ok := resp.(api.GetDeliveryQR200JSONResponse)
				require.True(t, ok, "expected 200 response, got %T", resp)
				require.NotEmpty(t, r.Token)
			case 403:
				r, ok := resp.(api.GetDeliveryQR403JSONResponse)
				require.True(t, ok, "expected 403 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 404:
				r, ok := resp.(api.GetDeliveryQR404JSONResponse)
				require.True(t, ok, "expected 404 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 409:
				r, ok := resp.(api.GetDeliveryQR409JSONResponse)
				require.True(t, ok, "expected 409 response, got %T", resp)
				require.Equal(t, tt.wantCode, string(r.Code))
			}
		})
	}
}

func TestHandler_VerifyDeliveryQR_Success(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	orderID := uuid.New()
	customerID := uuid.New()
	nonce := uuid.New()

	qrGen, err := qr.NewGenerator("test-secret-key-that-is-32-bytes-long!!", 24*time.Hour)
	require.NoError(t, err)

	delivery := &domain.Delivery{
		ID:      deliveryID,
		OrderID: orderID,
		UserID:  customerID,
		Status:  domain.DeliveryStatusReadyForPickup,
		QRNonce: uuid.NullUUID{UUID: nonce, Valid: true},
	}

	resp, err := qrGen.Generate(deliveryID, orderID, customerID, nonce)
	require.NoError(t, err)

	repo := &mocks.MockDeliveryRepository{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
			return delivery, nil
		},
		UpdateStatusConditionalFn: func(ctx context.Context, id uuid.UUID, fromStatus domain.DeliveryStatus, toStatus domain.DeliveryStatus, updates map[string]interface{}) (*domain.Delivery, error) {
			delivery.Status = domain.DeliveryStatusDelivered
			delivery.QRNonce = uuid.NullUUID{}
			return delivery, nil
		},
	}

	client := &mocks.MockOrderClient{
		UpdateOrderStatusFn: func(ctx context.Context, orderID uuid.UUID, status string) error {
			return nil
		},
	}

	handler := newTestHandler(repo, client, allowAllAuthorizer())
	req := api.VerifyDeliveryQRRequestObject{
		Body: &api.VerifyDeliveryQRJSONRequestBody{
			Token: resp.Token,
		},
	}

	claims := &auth.Claims{UserID: uuid.New(), Role: "employee"}
	ctx := auth.ContextWithClaims(context.Background(), claims)

	result, err := handler.VerifyDeliveryQR(ctx, req)
	require.NoError(t, err)

	r, ok := result.(api.VerifyDeliveryQR200JSONResponse)
	require.True(t, ok, "expected 200 response, got %T", result)
	require.Equal(t, true, r.Verified)
}

func TestHandler_VerifyDeliveryQR_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       string
		hasClaims  bool
		token      string
		auth       *mocks.MockAuthorizer
		wantStatus int
		wantCode   string
	}{
		{
			name:       "unauthorized - missing claims",
			role:       "employee",
			hasClaims:  false,
			token:      "some-token",
			auth:       allowAllAuthorizer(),
			wantStatus: 401,
			wantCode:   "unauthorized",
		},
		{
			name:      "forbidden - customer role",
			role:      "customer",
			hasClaims: true,
			token:     "some-token",
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					return rbac.ErrAccessDenied
				},
			},
			wantStatus: 403,
			wantCode:   "forbidden",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestHandler(&mocks.MockDeliveryRepository{}, &mocks.MockOrderClient{}, tt.auth)
			req := api.VerifyDeliveryQRRequestObject{
				Body: &api.VerifyDeliveryQRJSONRequestBody{
					Token: tt.token,
				},
			}

			var ctx context.Context
			if tt.hasClaims {
				claims := &auth.Claims{UserID: uuid.New(), Role: tt.role}
				ctx = auth.ContextWithClaims(context.Background(), claims)
			} else {
				ctx = context.Background()
			}

			result, err := handler.VerifyDeliveryQR(ctx, req)
			require.NoError(t, err)

			switch tt.wantStatus {
			case 401:
				r, ok := result.(api.VerifyDeliveryQR401JSONResponse)
				require.True(t, ok, "expected 401 response, got %T", result)
				require.Equal(t, tt.wantCode, string(r.Code))
			case 403:
				r, ok := result.(api.VerifyDeliveryQR403JSONResponse)
				require.True(t, ok, "expected 403 response, got %T", result)
				require.Equal(t, tt.wantCode, string(r.Code))
			}
		})
	}
}

func TestHandler_VerifyDeliveryQR_InvalidToken(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(&mocks.MockDeliveryRepository{}, &mocks.MockOrderClient{}, allowAllAuthorizer())
	req := api.VerifyDeliveryQRRequestObject{
		Body: &api.VerifyDeliveryQRJSONRequestBody{
			Token: "invalid.token.string",
		},
	}

	claims := &auth.Claims{UserID: uuid.New(), Role: "employee"}
	ctx := auth.ContextWithClaims(context.Background(), claims)

	result, err := handler.VerifyDeliveryQR(ctx, req)
	require.NoError(t, err)

	r, ok := result.(api.VerifyDeliveryQR400JSONResponse)
	require.True(t, ok, "expected 400 response, got %T", result)
	require.Equal(t, "invalid_token", string(r.Code))
}
