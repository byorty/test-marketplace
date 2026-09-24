package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	orderclient "github.com/byorty/test-marketplace/services/common/client/order"
	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/mocks"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/qr"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestService(repo *mocks.MockDeliveryRepository, orderClient *mocks.MockOrderClient, authorizer *mocks.MockAuthorizer) *DeliveryService {
	qrGen, err := qr.NewGenerator("test-secret-key-that-is-32-bytes-long!!", 24*time.Hour)
	if err != nil {
		panic(err)
	}
	qrVer, err := qr.NewVerifier("test-secret-key-that-is-32-bytes-long!!")
	if err != nil {
		panic(err)
	}
	return &DeliveryService{
		repo:        repo,
		log:         zap.NewNop(),
		orderClient: orderClient,
		qrGen:       qrGen,
		qrVer:       qrVer,
		validate:    validator.New(),
		authorizer:  authorizer,
		clock:       &RealClock{},
		notifier:    &LoggingNotifier{log: zap.NewNop()},
	}
}

func newTestServiceWithClock(repo *mocks.MockDeliveryRepository, orderClient *mocks.MockOrderClient, authorizer *mocks.MockAuthorizer, clock domain.Clock, notifier domain.Notifier) *DeliveryService {
	qrGen, err := qr.NewGenerator("test-secret-key-that-is-32-bytes-long!!", 24*time.Hour)
	if err != nil {
		panic(err)
	}
	qrVer, err := qr.NewVerifier("test-secret-key-that-is-32-bytes-long!!")
	if err != nil {
		panic(err)
	}
	return &DeliveryService{
		repo:        repo,
		log:         zap.NewNop(),
		orderClient: orderClient,
		qrGen:       qrGen,
		qrVer:       qrVer,
		validate:    validator.New(),
		authorizer:  authorizer,
		clock:       clock,
		notifier:    notifier,
	}
}

func allowAllAuthorizer() *mocks.MockAuthorizer {
	return &mocks.MockAuthorizer{
		AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
			return nil
		},
	}
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	orderID := uuid.New()
	userID := uuid.New()
	futureDate := time.Now().Add(48 * time.Hour)

	tests := []struct {
		name    string
		input   *domain.DeliveryCreateInput
		repo    *mocks.MockDeliveryRepository
		client  *mocks.MockOrderClient
		auth    *mocks.MockAuthorizer
		wantErr error
	}{
		{
			name: "success",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
				Role:                  "employee",
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
			auth: allowAllAuthorizer(),
		},
		{
			name: "customer cannot create delivery",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
				Role:                  "customer",
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
			wantErr: domain.ErrAccessDenied,
		},
		{
			name:    "nil input",
			input:   nil,
			repo:    &mocks.MockDeliveryRepository{},
			client:  &mocks.MockOrderClient{},
			auth:    allowAllAuthorizer(),
			wantErr: ErrNilInput,
		},
		{
			name: "nil order id",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               uuid.Nil,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
				Role:                  "employee",
			},
			repo:    &mocks.MockDeliveryRepository{},
			client:  &mocks.MockOrderClient{},
			auth:    allowAllAuthorizer(),
			wantErr: ErrInvalidOrderID,
		},
		{
			name: "nil user id",
			input: &domain.DeliveryCreateInput{
				UserID:                uuid.Nil,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
				Role:                  "employee",
			},
			repo:    &mocks.MockDeliveryRepository{},
			client:  &mocks.MockOrderClient{},
			auth:    allowAllAuthorizer(),
			wantErr: ErrInvalidUserID,
		},
		{
			name: "empty pickup address",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "",
				EstimatedDeliveryDate: futureDate,
				Role:                  "employee",
			},
			repo:    &mocks.MockDeliveryRepository{},
			client:  &mocks.MockOrderClient{},
			auth:    allowAllAuthorizer(),
			wantErr: ErrInvalidInput,
		},
		{
			name: "past estimated delivery date",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: time.Now().Add(-24 * time.Hour),
				Role:                  "employee",
			},
			repo:    &mocks.MockDeliveryRepository{},
			client:  &mocks.MockOrderClient{},
			auth:    allowAllAuthorizer(),
			wantErr: ErrInvalidEstimatedDate,
		},
		{
			name: "zero estimated delivery date",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: time.Time{},
				Role:                  "employee",
			},
			repo:    &mocks.MockDeliveryRepository{},
			client:  &mocks.MockOrderClient{},
			auth:    allowAllAuthorizer(),
			wantErr: ErrInvalidEstimatedDate,
		},
		{
			name: "order not paid",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
				Role:                  "employee",
			},
			repo: &mocks.MockDeliveryRepository{},
			client: &mocks.MockOrderClient{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*orderclient.OrderResponse, error) {
					return &orderclient.OrderResponse{ID: id, UserID: userID, Status: "CREATED"}, nil
				},
			},
			auth:    allowAllAuthorizer(),
			wantErr: domain.ErrOrderNotPaid,
		},
		{
			name: "order client error",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
				Role:                  "employee",
			},
			repo: &mocks.MockDeliveryRepository{},
			client: &mocks.MockOrderClient{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*orderclient.OrderResponse, error) {
					return nil, errors.New("order service unavailable")
				},
			},
			auth:    allowAllAuthorizer(),
			wantErr: errors.New("order service unavailable"),
		},
		{
			name: "delivery already exists for order",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
				Role:                  "employee",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByOrderIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     uuid.New(),
						Status: domain.DeliveryStatusInTransit,
					}, nil
				},
			},
			client: &mocks.MockOrderClient{
				GetOrderByIDFn: func(ctx context.Context, id uuid.UUID) (*orderclient.OrderResponse, error) {
					return &orderclient.OrderResponse{ID: id, UserID: userID, Status: "PAID"}, nil
				},
			},
			auth:    allowAllAuthorizer(),
			wantErr: domain.ErrDeliveryAlreadyExists,
		},
		{
			name: "repository create error",
			input: &domain.DeliveryCreateInput{
				UserID:                userID,
				OrderID:               orderID,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: futureDate,
				Role:                  "employee",
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
			auth:    allowAllAuthorizer(),
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, tt.client, tt.auth)
			result, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrDeliveryNotFound) ||
					errors.Is(tt.wantErr, domain.ErrOrderNotPaid) ||
					errors.Is(tt.wantErr, domain.ErrDeliveryAlreadyExists) ||
					errors.Is(tt.wantErr, ErrNilInput) ||
					errors.Is(tt.wantErr, ErrInvalidOrderID) ||
					errors.Is(tt.wantErr, ErrInvalidUserID) ||
					errors.Is(tt.wantErr, ErrInvalidEstimatedDate) ||
					errors.Is(tt.wantErr, domain.ErrAccessDenied) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.ErrorContains(t, err, tt.wantErr.Error())
				}
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotEqual(t, uuid.Nil, result.ID)
			require.Equal(t, domain.DeliveryStatusPending, result.Status)
			require.Equal(t, tt.input.OrderID, result.OrderID)
			require.Equal(t, tt.input.UserID, result.UserID)
			require.Equal(t, tt.input.PickupAddress, result.PickupAddress)
			require.False(t, result.CreatedAt.IsZero())
			require.False(t, result.UpdatedAt.IsZero())
		})
	}
}

func TestService_GetByID_Authorization(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	customerID := uuid.New()
	employeeID := uuid.New()
	otherCustomerID := uuid.New()

	newDelivery := func() *domain.Delivery {
		return &domain.Delivery{
			ID:     deliveryID,
			UserID: customerID,
			Status: domain.DeliveryStatusInTransit,
		}
	}

	tests := []struct {
		name    string
		userID  uuid.UUID
		role    string
		auth    *mocks.MockAuthorizer
		wantErr error
	}{
		{
			name:   "customer views own delivery",
			userID: customerID,
			role:   "customer",
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if role == "customer" && action == rbac.ActionView {
						return nil
					}
					return rbac.ErrAccessDenied
				},
			},
			wantErr: nil,
		},
		{
			name:   "customer views another customer's delivery - denied",
			userID: otherCustomerID,
			role:   "customer",
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					return rbac.ErrAccessDenied
				},
			},
			wantErr: domain.ErrAccessDenied,
		},
		{
			name:    "employee views any delivery",
			userID:  employeeID,
			role:    "employee",
			auth:    allowAllAuthorizer(),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := newDelivery()
			repo := &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return d, nil
				},
				ListReschedulesFn: func(ctx context.Context, deliveryID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			}

			svc := newTestService(repo, &mocks.MockOrderClient{}, tt.auth)
			result, err := svc.GetByID(context.Background(), tt.userID, deliveryID, tt.role)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	customerID := uuid.New()
	otherCustomerID := uuid.New()
	pastDate := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name    string
		userID  uuid.UUID
		role    string
		auth    *mocks.MockAuthorizer
		repo    *mocks.MockDeliveryRepository
		wantErr error
		wantNil bool
	}{
		{
			name:   "customer views own delivery with reschedules",
			userID: customerID,
			role:   "customer",
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
						OrderID:               uuid.New(),
						Status:                domain.DeliveryStatusInTransit,
						PickupAddress:         "123 Main St",
						EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
						IsLate:                false,
						CreatedAt:             time.Now(),
						UpdatedAt:             time.Now(),
					}, nil
				},
				ListReschedulesFn: func(ctx context.Context, dID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return []domain.DeliveryReschedule{
						{
							ID:           uuid.New(),
							DeliveryID:   deliveryID,
							PreviousDate: pastDate.Add(48 * time.Hour),
							NewDate:      time.Now().Add(24 * time.Hour),
							Reason:       "weather delay",
							CreatedAt:    time.Now(),
						},
					}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
		},
		{
			name:   "employee views any delivery",
			userID: uuid.New(),
			role:   "employee",
			auth:   allowAllAuthorizer(),
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     deliveryID,
						UserID: customerID,
						Status: domain.DeliveryStatusPending,
					}, nil
				},
				ListReschedulesFn: func(ctx context.Context, dID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
		},
		{
			name:   "IDOR: customer B cannot view customer A delivery",
			userID: otherCustomerID,
			role:   "customer",
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
			wantErr: domain.ErrAccessDenied,
			wantNil: true,
		},
		{
			name:   "delivery not found",
			userID: customerID,
			role:   "customer",
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
			wantErr: domain.ErrDeliveryNotFound,
			wantNil: true,
		},
		{
			name:    "invalid id",
			userID:  customerID,
			role:    "customer",
			auth:    allowAllAuthorizer(),
			repo:    &mocks.MockDeliveryRepository{},
			wantErr: ErrInvalidID,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id := deliveryID
			if tt.name == "invalid id" {
				id = uuid.Nil
			}

			svc := newTestService(tt.repo, &mocks.MockOrderClient{}, tt.auth)
			result, err := svc.GetByID(context.Background(), tt.userID, id, tt.role)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				if tt.wantNil {
					require.Nil(t, result)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, deliveryID, result.ID)
		})
	}
}

func TestService_Reschedule_Authorization(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	futureDate := time.Now().Add(72 * time.Hour)

	tests := []struct {
		name    string
		role    string
		auth    *mocks.MockAuthorizer
		wantErr error
	}{
		{
			name:    "employee can reschedule",
			role:    "employee",
			auth:    allowAllAuthorizer(),
			wantErr: nil,
		},
		{
			name: "customer cannot reschedule",
			role: "customer",
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if role == "customer" && action == rbac.ActionReschedule {
						return rbac.ErrAccessDenied
					}
					return nil
				},
			},
			wantErr: domain.ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:                    deliveryID,
						Status:                domain.DeliveryStatusPending,
						EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
					}, nil
				},
				RescheduleTxFn: func(ctx context.Context, d *domain.Delivery, r *domain.DeliveryReschedule) (*domain.Delivery, error) {
					return d, nil
				},
				ListReschedulesFn: func(ctx context.Context, deliveryID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
			}

			svc := newTestService(repo, &mocks.MockOrderClient{}, tt.auth)
			result, err := svc.Reschedule(context.Background(), uuid.New(), deliveryID, tt.role, &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "delay",
			})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestService_UpdateStatus_Authorization(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()

	tests := []struct {
		name    string
		role    string
		auth    *mocks.MockAuthorizer
		wantErr error
	}{
		{
			name:    "employee can update status",
			role:    "employee",
			auth:    allowAllAuthorizer(),
			wantErr: nil,
		},
		{
			name: "customer cannot update status",
			role: "customer",
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if role == "customer" && action == rbac.ActionUpdateStatus {
						return rbac.ErrAccessDenied
					}
					return nil
				},
			},
			wantErr: domain.ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			}

			svc := newTestService(repo, &mocks.MockOrderClient{}, tt.auth)
			result, err := svc.UpdateStatus(context.Background(), uuid.New(), deliveryID, tt.role, domain.DeliveryStatusInTransit)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestService_UpdateStatus(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()

	tests := []struct {
		name      string
		delivery  *domain.Delivery
		newStatus domain.DeliveryStatus
		repo      *mocks.MockDeliveryRepository
		wantErr   error
	}{
		{
			name: "pending -> in_transit",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusPending,
			},
			newStatus: domain.DeliveryStatusInTransit,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
		},
		{
			name: "in_transit -> ready_for_pickup",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusInTransit,
			},
			newStatus: domain.DeliveryStatusReadyForPickup,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusInTransit}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
		},
		{
			name: "ready_for_pickup -> delivered",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusReadyForPickup,
			},
			newStatus: domain.DeliveryStatusDelivered,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusReadyForPickup}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
		},
		{
			name: "pending -> cancelled",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusPending,
			},
			newStatus: domain.DeliveryStatusCancelled,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
		},
		{
			name: "in_transit -> cancelled",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusInTransit,
			},
			newStatus: domain.DeliveryStatusCancelled,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusInTransit}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
		},
		{
			name: "ready_for_pickup -> cancelled",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusReadyForPickup,
			},
			newStatus: domain.DeliveryStatusCancelled,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusReadyForPickup}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
		},
		{
			name:      "invalid id",
			delivery:  nil,
			newStatus: domain.DeliveryStatusInTransit,
			repo:      &mocks.MockDeliveryRepository{},
			wantErr:   ErrInvalidID,
		},
		{
			name:      "invalid status value",
			delivery:  &domain.Delivery{ID: deliveryID},
			newStatus: domain.DeliveryStatus("INVALID"),
			repo:      &mocks.MockDeliveryRepository{},
			wantErr:   ErrInvalidStatus,
		},
		{
			name: "delivery not found",
			delivery: &domain.Delivery{
				ID: deliveryID,
			},
			newStatus: domain.DeliveryStatusInTransit,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
			},
			wantErr: domain.ErrDeliveryNotFound,
		},
		{
			name: "forbidden transition: pending -> delivered",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusPending,
			},
			newStatus: domain.DeliveryStatusDelivered,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
				},
			},
			wantErr: domain.ErrInvalidStatusTransition,
		},
		{
			name: "forbidden transition: pending -> ready_for_pickup",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusPending,
			},
			newStatus: domain.DeliveryStatusReadyForPickup,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
				},
			},
			wantErr: domain.ErrInvalidStatusTransition,
		},
		{
			name: "forbidden transition: pending -> pending (same status)",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusPending,
			},
			newStatus: domain.DeliveryStatusPending,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
				},
			},
			wantErr: domain.ErrInvalidStatusTransition,
		},
		{
			name: "forbidden transition: in_transit -> pending",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusInTransit,
			},
			newStatus: domain.DeliveryStatusPending,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusInTransit}, nil
				},
			},
			wantErr: domain.ErrInvalidStatusTransition,
		},
		{
			name: "forbidden transition: in_transit -> delivered",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusInTransit,
			},
			newStatus: domain.DeliveryStatusDelivered,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusInTransit}, nil
				},
			},
			wantErr: domain.ErrInvalidStatusTransition,
		},
		{
			name: "forbidden transition: ready_for_pickup -> in_transit",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusReadyForPickup,
			},
			newStatus: domain.DeliveryStatusInTransit,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusReadyForPickup}, nil
				},
			},
			wantErr: domain.ErrInvalidStatusTransition,
		},
		{
			name: "forbidden transition: ready_for_pickup -> pending",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusReadyForPickup,
			},
			newStatus: domain.DeliveryStatusPending,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusReadyForPickup}, nil
				},
			},
			wantErr: domain.ErrInvalidStatusTransition,
		},
		{
			name: "delivered is terminal - any transition denied",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusDelivered,
			},
			newStatus: domain.DeliveryStatusInTransit,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusDelivered}, nil
				},
			},
			wantErr: domain.ErrDeliveryAlreadyDelivered,
		},
		{
			name: "cancelled is terminal - any transition denied",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusCancelled,
			},
			newStatus: domain.DeliveryStatusInTransit,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusCancelled}, nil
				},
			},
			wantErr: domain.ErrDeliveryAlreadyCancelled,
		},
		{
			name: "delivered -> cancelled denied (terminal)",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusDelivered,
			},
			newStatus: domain.DeliveryStatusCancelled,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusDelivered}, nil
				},
			},
			wantErr: domain.ErrDeliveryAlreadyDelivered,
		},
		{
			name: "cancelled -> delivered denied (terminal)",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusCancelled,
			},
			newStatus: domain.DeliveryStatusDelivered,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusCancelled}, nil
				},
			},
			wantErr: domain.ErrDeliveryAlreadyCancelled,
		},
		{
			name: "persistence failure",
			delivery: &domain.Delivery{
				ID:     deliveryID,
				Status: domain.DeliveryStatusPending,
			},
			newStatus: domain.DeliveryStatusInTransit,
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return nil, errors.New("db error: update failed")
				},
			},
			wantErr: errors.New("update delivery"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, &mocks.MockOrderClient{}, allowAllAuthorizer())

			deliveryIDArg := uuid.Nil
			if tt.delivery != nil {
				deliveryIDArg = tt.delivery.ID
			}

			result, err := svc.UpdateStatus(context.Background(), uuid.New(), deliveryIDArg, "employee", tt.newStatus)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrDeliveryNotFound) ||
					errors.Is(tt.wantErr, domain.ErrInvalidStatusTransition) ||
					errors.Is(tt.wantErr, domain.ErrDeliveryAlreadyDelivered) ||
					errors.Is(tt.wantErr, domain.ErrDeliveryAlreadyCancelled) ||
					errors.Is(tt.wantErr, ErrInvalidID) ||
					errors.Is(tt.wantErr, ErrInvalidStatus) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.ErrorContains(t, err, tt.wantErr.Error())
				}
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tt.newStatus, result.Status)
		})
	}
}

func TestService_Reschedule(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	futureDate := time.Now().Add(72 * time.Hour)

	tests := []struct {
		name    string
		input   *domain.RescheduleInput
		repo    *mocks.MockDeliveryRepository
		wantErr error
		check   func(t *testing.T, result *domain.Delivery)
	}{
		{
			name: "success - pending delivery",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "delayed shipment",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:                    deliveryID,
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
		},
		{
			name: "success - in_transit delivery",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "traffic delay",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:                    deliveryID,
						Status:                domain.DeliveryStatusInTransit,
						EstimatedDeliveryDate: time.Now().Add(12 * time.Hour),
					}, nil
				},
				RescheduleTxFn: func(ctx context.Context, d *domain.Delivery, r *domain.DeliveryReschedule) (*domain.Delivery, error) {
					return d, nil
				},
				ListReschedulesFn: func(ctx context.Context, dID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
			},
		},
		{
			name: "success - previous_date matches current estimated_delivery_date",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "weather",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:                    deliveryID,
						Status:                domain.DeliveryStatusPending,
						EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
					}, nil
				},
				RescheduleTxFn: func(ctx context.Context, d *domain.Delivery, r *domain.DeliveryReschedule) (*domain.Delivery, error) {
					require.NotNil(t, r)
					require.Equal(t, deliveryID, r.DeliveryID)
					require.False(t, r.PreviousDate.IsZero())
					require.Equal(t, futureDate.Truncate(time.Second), r.NewDate.Truncate(time.Second))
					require.Equal(t, "weather", r.Reason)
					return d, nil
				},
				ListReschedulesFn: func(ctx context.Context, dID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
			},
		},
		{
			name:    "nil input",
			input:   nil,
			repo:    &mocks.MockDeliveryRepository{},
			wantErr: ErrNilInput,
		},
		{
			name: "reschedule on ready_for_pickup denied",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "too late",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     deliveryID,
						Status: domain.DeliveryStatusReadyForPickup,
					}, nil
				},
			},
			wantErr: domain.ErrRescheduleOnTerminal,
		},
		{
			name: "reschedule on delivered denied",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "too late",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     deliveryID,
						Status: domain.DeliveryStatusDelivered,
					}, nil
				},
			},
			wantErr: domain.ErrRescheduleOnTerminal,
		},
		{
			name: "reschedule on cancelled denied",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "too late",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:     deliveryID,
						Status: domain.DeliveryStatusCancelled,
					}, nil
				},
			},
			wantErr: domain.ErrRescheduleOnTerminal,
		},
		{
			name: "past new date denied",
			input: &domain.RescheduleInput{
				NewDate: time.Now().Add(-24 * time.Hour),
				Reason:  "past date",
			},
			repo:    &mocks.MockDeliveryRepository{},
			wantErr: ErrInvalidNewDate,
		},
		{
			name: "zero new date denied",
			input: &domain.RescheduleInput{
				NewDate: time.Time{},
				Reason:  "zero date",
			},
			repo:    &mocks.MockDeliveryRepository{},
			wantErr: ErrInvalidNewDate,
		},
		{
			name: "delivery not found",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "not found",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
			},
			wantErr: domain.ErrDeliveryNotFound,
		},
		{
			name: "invalid delivery id",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "invalid id",
			},
			repo:    &mocks.MockDeliveryRepository{},
			wantErr: ErrInvalidID,
		},
		{
			name: "reschedule transaction failure",
			input: &domain.RescheduleInput{
				NewDate: futureDate,
				Reason:  "db error",
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return &domain.Delivery{
						ID:                    deliveryID,
						Status:                domain.DeliveryStatusPending,
						EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
					}, nil
				},
				RescheduleTxFn: func(ctx context.Context, d *domain.Delivery, r *domain.DeliveryReschedule) (*domain.Delivery, error) {
					return nil, errors.New("db error: transaction failed")
				},
			},
			wantErr: errors.New("reschedule delivery"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, &mocks.MockOrderClient{}, allowAllAuthorizer())

			deliveryIDArg := deliveryID
			if tt.name == "invalid delivery id" {
				deliveryIDArg = uuid.Nil
			}

			result, err := svc.Reschedule(context.Background(), uuid.New(), deliveryIDArg, "employee", tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrDeliveryNotFound) ||
					errors.Is(tt.wantErr, domain.ErrRescheduleOnTerminal) ||
					errors.Is(tt.wantErr, ErrNilInput) ||
					errors.Is(tt.wantErr, ErrInvalidNewDate) ||
					errors.Is(tt.wantErr, ErrInvalidID) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.ErrorContains(t, err, tt.wantErr.Error())
				}
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestService_GetQR_Authorization(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	customerID := uuid.New()
	otherCustomerID := uuid.New()
	employeeID := uuid.New()

	newDelivery := func() *domain.Delivery {
		return &domain.Delivery{
			ID:     deliveryID,
			UserID: customerID,
			Status: domain.DeliveryStatusReadyForPickup,
		}
	}

	tests := []struct {
		name    string
		userID  uuid.UUID
		role    string
		auth    *mocks.MockAuthorizer
		wantErr error
	}{
		{
			name:   "customer gets own QR",
			userID: customerID,
			role:   "customer",
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					if role == "customer" && action == rbac.ActionGetQR {
						return nil
					}
					if role == "customer" && action == rbac.ActionVerifyQR {
						return rbac.ErrAccessDenied
					}
					return nil
				},
			},
			wantErr: nil,
		},
		{
			name:   "customer gets another customer's QR - denied",
			userID: otherCustomerID,
			role:   "customer",
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					return rbac.ErrAccessDenied
				},
			},
			wantErr: domain.ErrAccessDenied,
		},
		{
			name:    "employee gets any QR",
			userID:  employeeID,
			role:    "employee",
			auth:    allowAllAuthorizer(),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := newDelivery()
			repo := &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return d, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			}

			svc := newTestService(repo, &mocks.MockOrderClient{}, tt.auth)
			result, err := svc.GetQR(context.Background(), tt.userID, deliveryID, tt.role)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestService_GetByID_OverdueNotification(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	pastDate := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name       string
		delivery   *domain.Delivery
		now        time.Time
		wantLate   bool
		wantNotify bool
	}{
		{
			name: "delivery becomes overdue - notifies",
			delivery: &domain.Delivery{
				ID:                    deliveryID,
				Status:                domain.DeliveryStatusInTransit,
				EstimatedDeliveryDate: pastDate,
				IsLate:                false,
			},
			now:        time.Now(),
			wantLate:   true,
			wantNotify: true,
		},
		{
			name: "delivery already overdue - no duplicate notification",
			delivery: &domain.Delivery{
				ID:                    deliveryID,
				Status:                domain.DeliveryStatusInTransit,
				EstimatedDeliveryDate: pastDate,
				IsLate:                true,
			},
			now:        time.Now(),
			wantLate:   true,
			wantNotify: false,
		},
		{
			name: "delivery on time - no notification",
			delivery: &domain.Delivery{
				ID:                    deliveryID,
				Status:                domain.DeliveryStatusPending,
				EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
				IsLate:                false,
			},
			now:        time.Now(),
			wantLate:   false,
			wantNotify: false,
		},
		{
			name: "terminal status delivered - not late, no notification",
			delivery: &domain.Delivery{
				ID:                    deliveryID,
				Status:                domain.DeliveryStatusDelivered,
				EstimatedDeliveryDate: pastDate,
				IsLate:                true,
			},
			now:        time.Now(),
			wantLate:   false,
			wantNotify: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			notifier := &mocks.MockNotifier{}
			notifyCalled := false
			notifier.NotifyDeliveryLateFn = func(ctx context.Context, d *domain.Delivery) error {
				notifyCalled = true
				return nil
			}

			clock := &mocks.MockClock{NowTime: tt.now}

			repo := &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return tt.delivery, nil
				},
				ListReschedulesFn: func(ctx context.Context, dID uuid.UUID) ([]domain.DeliveryReschedule, error) {
					return nil, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			}

			svc := newTestServiceWithClock(repo, &mocks.MockOrderClient{}, allowAllAuthorizer(), clock, notifier)
			result, err := svc.GetByID(context.Background(), uuid.New(), deliveryID, "employee")

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tt.wantLate, result.IsLate)
			require.Equal(t, tt.wantNotify, notifyCalled)
		})
	}
}

func TestService_UpdateStatus_Notification(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()

	notifier := &mocks.MockNotifier{}
	notifyCalled := false
	notifier.NotifyDeliveryStatusChangedFn = func(ctx context.Context, d *domain.Delivery, oldStatus domain.DeliveryStatus) error {
		notifyCalled = true
		require.Equal(t, domain.DeliveryStatusPending, oldStatus)
		return nil
	}

	repo := &mocks.MockDeliveryRepository{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
			return &domain.Delivery{ID: deliveryID, Status: domain.DeliveryStatusPending}, nil
		},
		UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
			return d, nil
		},
	}

	svc := newTestServiceWithClock(repo, &mocks.MockOrderClient{}, allowAllAuthorizer(), &RealClock{}, notifier)
	result, err := svc.UpdateStatus(context.Background(), uuid.New(), deliveryID, "employee", domain.DeliveryStatusInTransit)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, notifyCalled)
}

func TestService_GetQR(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	orderID := uuid.New()
	customerID := uuid.New()

	delivery := &domain.Delivery{
		ID:      deliveryID,
		OrderID: orderID,
		UserID:  customerID,
		Status:  domain.DeliveryStatusReadyForPickup,
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		deliveryID uuid.UUID
		role       string
		auth       *mocks.MockAuthorizer
		repo       *mocks.MockDeliveryRepository
		wantErr    error
	}{
		{
			name:       "customer gets own QR - success",
			userID:     customerID,
			deliveryID: deliveryID,
			role:       "customer",
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
					return delivery, nil
				},
				UpdateFn: func(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
					return d, nil
				},
			},
			wantErr: nil,
		},
		{
			name:       "customer gets another's QR - access denied",
			userID:     uuid.New(),
			deliveryID: deliveryID,
			role:       "customer",
			auth: &mocks.MockAuthorizer{
				AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
					return rbac.ErrAccessDenied
				},
			},
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return delivery, nil
				},
			},
			wantErr: domain.ErrAccessDenied,
		},
		{
			name:       "delivery not found",
			userID:     customerID,
			deliveryID: deliveryID,
			role:       "customer",
			auth:       allowAllAuthorizer(),
			repo: &mocks.MockDeliveryRepository{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
					return nil, domain.ErrDeliveryNotFound
				},
			},
			wantErr: domain.ErrDeliveryNotFound,
		},
		{
			name:       "wrong status - QR not available",
			userID:     customerID,
			deliveryID: deliveryID,
			role:       "customer",
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
			wantErr: domain.ErrQRNotAvailable,
		},
		{
			name:       "invalid delivery ID",
			userID:     customerID,
			deliveryID: uuid.Nil,
			role:       "customer",
			auth:       allowAllAuthorizer(),
			repo:       &mocks.MockDeliveryRepository{},
			wantErr:    ErrInvalidID,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(tt.repo, &mocks.MockOrderClient{}, tt.auth)
			result, err := svc.GetQR(context.Background(), tt.userID, tt.deliveryID, tt.role)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotEmpty(t, result.Token)
			require.False(t, result.ExpiresAt.IsZero())
		})
	}
}

func TestService_VerifyQR(t *testing.T) {
	t.Parallel()

	deliveryID := uuid.New()
	orderID := uuid.New()
	customerID := uuid.New()
	otherUserID := uuid.New()

	qrGen, err := qr.NewGenerator("test-secret-key-that-is-32-bytes-long!!", 24*time.Hour)
	require.NoError(t, err)

	wrongKeyGen, err := qr.NewGenerator(strings.Repeat("x", 32), 24*time.Hour)
	require.NoError(t, err)

	t.Run("successful verification", func(t *testing.T) {
		t.Parallel()

		nonce := uuid.New()
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

		orderClient := &mocks.MockOrderClient{
			UpdateOrderStatusFn: func(ctx context.Context, orderID uuid.UUID, status string) error {
				return nil
			},
		}

		svc := newTestService(repo, orderClient, allowAllAuthorizer())
		result, err := svc.VerifyQR(context.Background(), resp.Token)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, domain.DeliveryStatusDelivered, result.Status)
	})

	t.Run("invalid token - bad signature", func(t *testing.T) {
		t.Parallel()

		svc := newTestService(&mocks.MockDeliveryRepository{}, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err := svc.VerifyQR(context.Background(), "invalid.token.data")
		require.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()

		pastTime := time.Now().Add(-2 * time.Hour)
		mockClock := &mocks.MockClock{NowTime: pastTime}
		expiredGen, err := qr.NewGeneratorWithClock("test-secret-key-that-is-32-bytes-long!!", 1*time.Second, mockClock)
		require.NoError(t, err)

		nonce := uuid.New()
		resp, err := expiredGen.Generate(deliveryID, orderID, customerID, nonce)
		require.NoError(t, err)

		svc := newTestService(&mocks.MockDeliveryRepository{}, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err = svc.VerifyQR(context.Background(), resp.Token)
		require.ErrorIs(t, err, qr.ErrTokenExpired)
	})

	t.Run("revoked token - nonce mismatch", func(t *testing.T) {
		t.Parallel()

		oldNonce := uuid.New()
		newNonce := uuid.New()
		delivery := &domain.Delivery{
			ID:      deliveryID,
			OrderID: orderID,
			UserID:  customerID,
			Status:  domain.DeliveryStatusReadyForPickup,
			QRNonce: uuid.NullUUID{UUID: newNonce, Valid: true},
		}

		resp, err := qrGen.Generate(deliveryID, orderID, customerID, oldNonce)
		require.NoError(t, err)

		repo := &mocks.MockDeliveryRepository{
			GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
				return delivery, nil
			},
		}

		svc := newTestService(repo, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err = svc.VerifyQR(context.Background(), resp.Token)
		require.ErrorIs(t, err, qr.ErrTokenRevoked)
	})

	t.Run("revoked token - no nonce in DB", func(t *testing.T) {
		t.Parallel()

		nonce := uuid.New()
		delivery := &domain.Delivery{
			ID:      deliveryID,
			OrderID: orderID,
			UserID:  customerID,
			Status:  domain.DeliveryStatusReadyForPickup,
			QRNonce: uuid.NullUUID{},
		}

		resp, err := qrGen.Generate(deliveryID, orderID, customerID, nonce)
		require.NoError(t, err)

		repo := &mocks.MockDeliveryRepository{
			GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
				return delivery, nil
			},
		}

		svc := newTestService(repo, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err = svc.VerifyQR(context.Background(), resp.Token)
		require.ErrorIs(t, err, qr.ErrTokenRevoked)
	})

	t.Run("wrong status - already delivered", func(t *testing.T) {
		t.Parallel()

		nonce := uuid.New()
		delivery := &domain.Delivery{
			ID:      deliveryID,
			OrderID: orderID,
			UserID:  customerID,
			Status:  domain.DeliveryStatusDelivered,
			QRNonce: uuid.NullUUID{UUID: nonce, Valid: true},
		}

		resp, err := qrGen.Generate(deliveryID, orderID, customerID, nonce)
		require.NoError(t, err)

		repo := &mocks.MockDeliveryRepository{
			GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
				return delivery, nil
			},
		}

		svc := newTestService(repo, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err = svc.VerifyQR(context.Background(), resp.Token)
		require.ErrorIs(t, err, domain.ErrQRNotAvailable)
	})

	t.Run("wrong status - in transit", func(t *testing.T) {
		t.Parallel()

		nonce := uuid.New()
		delivery := &domain.Delivery{
			ID:      deliveryID,
			OrderID: orderID,
			UserID:  customerID,
			Status:  domain.DeliveryStatusInTransit,
			QRNonce: uuid.NullUUID{UUID: nonce, Valid: true},
		}

		resp, err := qrGen.Generate(deliveryID, orderID, customerID, nonce)
		require.NoError(t, err)

		repo := &mocks.MockDeliveryRepository{
			GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
				return delivery, nil
			},
		}

		svc := newTestService(repo, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err = svc.VerifyQR(context.Background(), resp.Token)
		require.ErrorIs(t, err, domain.ErrQRNotAvailable)
	})

	t.Run("token for different user - access denied", func(t *testing.T) {
		t.Parallel()

		nonce := uuid.New()
		delivery := &domain.Delivery{
			ID:      deliveryID,
			OrderID: orderID,
			UserID:  customerID,
			Status:  domain.DeliveryStatusReadyForPickup,
			QRNonce: uuid.NullUUID{UUID: nonce, Valid: true},
		}

		resp, err := qrGen.Generate(deliveryID, orderID, otherUserID, nonce)
		require.NoError(t, err)

		repo := &mocks.MockDeliveryRepository{
			GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
				return delivery, nil
			},
		}

		svc := newTestService(repo, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err = svc.VerifyQR(context.Background(), resp.Token)
		require.ErrorIs(t, err, domain.ErrAccessDenied)
	})

	t.Run("delivery not found", func(t *testing.T) {
		t.Parallel()

		nonce := uuid.New()
		resp, err := qrGen.Generate(deliveryID, orderID, customerID, nonce)
		require.NoError(t, err)

		repo := &mocks.MockDeliveryRepository{
			GetByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
				return nil, domain.ErrDeliveryNotFound
			},
		}

		svc := newTestService(repo, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err = svc.VerifyQR(context.Background(), resp.Token)
		require.ErrorIs(t, err, domain.ErrDeliveryNotFound)
	})

	t.Run("modified signature - rejected", func(t *testing.T) {
		t.Parallel()

		nonce := uuid.New()
		resp, err := wrongKeyGen.Generate(deliveryID, orderID, customerID, nonce)
		require.NoError(t, err)

		svc := newTestService(&mocks.MockDeliveryRepository{}, &mocks.MockOrderClient{}, allowAllAuthorizer())
		_, err = svc.VerifyQR(context.Background(), resp.Token)
		require.ErrorIs(t, err, qr.ErrInvalidSignature)
	})
}
