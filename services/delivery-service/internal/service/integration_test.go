package service

import (
	"context"
	"testing"
	"time"

	orderclient "github.com/byorty/test-marketplace/services/common/client/order"
	"github.com/byorty/test-marketplace/services/common/rbac"
	testtools "github.com/byorty/test-marketplace/services/common/test-tools"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/mocks"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/qr"
	deliveryrepo "github.com/byorty/test-marketplace/services/delivery-service/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

func newIntegrationService(t *testing.T) (*DeliveryService, *bun.DB) {
	t.Helper()

	db := testtools.NewTestDB(t)
	repo := deliveryrepo.New(db, zap.NewNop())

	qrGen, err := qr.NewGenerator("integration-test-secret-key-32-bytes!!", 24*time.Hour)
	require.NoError(t, err)

	qrVer, err := qr.NewVerifier("integration-test-secret-key-32-bytes!!")
	require.NoError(t, err)

	mockOrderClient := &mocks.MockOrderClient{
		GetOrderByIDFn: func(ctx context.Context, orderID uuid.UUID) (*orderclient.OrderResponse, error) {
			return &orderclient.OrderResponse{
				ID:     orderID,
				UserID: uuid.New(),
				Status: "PAID",
			}, nil
		},
		UpdateOrderStatusFn: func(ctx context.Context, orderID uuid.UUID, status string) error {
			return nil
		},
	}

	authorizer := &mocks.MockAuthorizer{
		AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
			return nil
		},
	}

	mockNotifier := &mocks.MockNotifier{
		NotifyDeliveryLateFn: func(ctx context.Context, d *domain.Delivery) error {
			return nil
		},
		NotifyDeliveryStatusChangedFn: func(ctx context.Context, d *domain.Delivery, oldStatus domain.DeliveryStatus) error {
			return nil
		},
	}

	svc := New(
		zap.NewNop(),
		repo,
		mockOrderClient,
		qrGen,
		qrVer,
		validator.New(),
		authorizer,
		&RealClock{},
		mockNotifier,
	)

	return svc, db
}

func createTestDelivery(t *testing.T, svc *DeliveryService, userID uuid.UUID) *domain.Delivery {
	t.Helper()

	input := &domain.DeliveryCreateInput{
		UserID:                userID,
		OrderID:               uuid.New(),
		PickupAddress:         "123 Test St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		Role:                  "employee",
	}

	delivery, err := svc.Create(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, delivery)

	return delivery
}

// Scenario 1: Employee creates a delivery
func TestIntegration_CreateDelivery(t *testing.T) {
	t.Parallel()

	svc, db := newIntegrationService(t)
	_ = db

	userID := uuid.New()
	orderID := uuid.New()

	input := &domain.DeliveryCreateInput{
		UserID:                userID,
		OrderID:               orderID,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		Role:                  "employee",
	}

	delivery, err := svc.Create(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, delivery)
	require.Equal(t, domain.DeliveryStatusPending, delivery.Status)
	require.Equal(t, orderID, delivery.OrderID)
	require.Equal(t, userID, delivery.UserID)
	require.Equal(t, "123 Main St", delivery.PickupAddress)
	require.False(t, delivery.CreatedAt.IsZero())
	require.False(t, delivery.UpdatedAt.IsZero())
}

// Scenario 2: Customer gets their own delivery
func TestIntegration_CustomerGetsOwnDelivery(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	found, err := svc.GetByID(context.Background(), userID, delivery.ID, "customer")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, delivery.ID, found.ID)
	require.Equal(t, userID, found.UserID)
}

// Scenario 3: Customer tries to get another's delivery — access denied
func TestIntegration_CustomerGetsOtherDelivery_AccessDenied(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	ownerID := uuid.New()
	otherID := uuid.New()
	delivery := createTestDelivery(t, svc, ownerID)

	_, err := svc.GetByID(context.Background(), otherID, delivery.ID, "customer")
	require.ErrorIs(t, err, domain.ErrAccessDenied)
}

// Scenario 4: Employee reschedules a delivery
func TestIntegration_RescheduleDelivery(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	newDate := time.Now().Add(72 * time.Hour)
	input := &domain.RescheduleInput{
		NewDate: newDate,
		Reason:  "weather delay",
	}

	rescheduled, err := svc.Reschedule(context.Background(), userID, delivery.ID, "employee", input)
	require.NoError(t, err)
	require.NotNil(t, rescheduled)
	require.Equal(t, newDate.UTC().Truncate(time.Millisecond), rescheduled.EstimatedDeliveryDate.UTC().Truncate(time.Millisecond))
	require.False(t, rescheduled.IsLate)
}

// Scenario 5: Customer tries to reschedule — access denied
func TestIntegration_CustomerReschedule_AccessDenied(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	denyAllAuthorizer := &mocks.MockAuthorizer{
		AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error {
			if role == "customer" && action == rbac.ActionReschedule {
				return rbac.ErrAccessDenied
			}
			return nil
		},
	}

	denySvc := New(
		zap.NewNop(),
		newRepoFromService(t, svc),
		&mocks.MockOrderClient{
			GetOrderByIDFn: func(ctx context.Context, orderID uuid.UUID) (*orderclient.OrderResponse, error) {
				return &orderclient.OrderResponse{ID: orderID, UserID: uuid.New(), Status: "PAID"}, nil
			},
		},
		mustNewGenerator(t),
		mustNewVerifier(t),
		validator.New(),
		denyAllAuthorizer,
		&RealClock{},
		&mocks.MockNotifier{},
	)

	input := &domain.RescheduleInput{
		NewDate: time.Now().Add(72 * time.Hour),
		Reason:  "test",
	}

	_, err := denySvc.Reschedule(context.Background(), userID, delivery.ID, "customer", input)
	require.ErrorIs(t, err, domain.ErrAccessDenied)
}

// Scenario 6: Employee changes status through valid transitions
func TestIntegration_StatusTransitions_Valid(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	updated, err := svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusInTransit)
	require.NoError(t, err)
	require.Equal(t, domain.DeliveryStatusInTransit, updated.Status)

	updated, err = svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusReadyForPickup)
	require.NoError(t, err)
	require.Equal(t, domain.DeliveryStatusReadyForPickup, updated.Status)

	updated, err = svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusDelivered)
	require.NoError(t, err)
	require.Equal(t, domain.DeliveryStatusDelivered, updated.Status)
}

// Scenario 7: Employee attempts forbidden status transition
func TestIntegration_StatusTransitions_Forbidden(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	_, err := svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusDelivered)
	require.ErrorIs(t, err, domain.ErrInvalidStatusTransition)

	_, err = svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusReadyForPickup)
	require.ErrorIs(t, err, domain.ErrInvalidStatusTransition)
}

// Scenario 8: Delivery becomes overdue
func TestIntegration_OverdueDelivery(t *testing.T) {
	t.Parallel()

	db := testtools.NewTestDB(t)
	repo := deliveryrepo.New(db, zap.NewNop())

	pastDate := time.Now().Add(-24 * time.Hour)
	userID := uuid.New()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                userID,
		Status:                domain.DeliveryStatusInTransit,
		PickupAddress:         "123 Overdue St",
		EstimatedDeliveryDate: pastDate,
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(context.Background(), delivery)
	require.NoError(t, err)

	mockClock := &mocks.MockClock{NowTime: time.Now()}
	lateNotifier := &mocks.MockNotifier{}
	notifyCalled := false
	lateNotifier.NotifyDeliveryLateFn = func(ctx context.Context, d *domain.Delivery) error {
		notifyCalled = true
		return nil
	}

	qrGen, err := qr.NewGenerator("integration-test-secret-key-32-bytes!!", 24*time.Hour)
	require.NoError(t, err)
	qrVer, err := qr.NewVerifier("integration-test-secret-key-32-bytes!!")
	require.NoError(t, err)

	svc := New(
		zap.NewNop(),
		repo,
		&mocks.MockOrderClient{
			GetOrderByIDFn: func(ctx context.Context, orderID uuid.UUID) (*orderclient.OrderResponse, error) {
				return &orderclient.OrderResponse{ID: orderID, UserID: uuid.New(), Status: "PAID"}, nil
			},
		},
		qrGen,
		qrVer,
		validator.New(),
		&mocks.MockAuthorizer{AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error { return nil }},
		mockClock,
		lateNotifier,
	)

	found, err := svc.GetByID(context.Background(), userID, delivery.ID, "employee")
	require.NoError(t, err)
	require.True(t, found.IsLate)
	require.True(t, notifyCalled)
}

// Scenario 9: Customer gets info about overdue delivery
func TestIntegration_CustomerGetsOverdueDeliveryInfo(t *testing.T) {
	t.Parallel()

	db := testtools.NewTestDB(t)
	repo := deliveryrepo.New(db, zap.NewNop())

	pastDate := time.Now().Add(-24 * time.Hour)
	userID := uuid.New()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                userID,
		Status:                domain.DeliveryStatusInTransit,
		PickupAddress:         "123 Overdue St",
		EstimatedDeliveryDate: pastDate,
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(context.Background(), delivery)
	require.NoError(t, err)

	mockClock := &mocks.MockClock{NowTime: time.Now()}

	svc := New(
		zap.NewNop(),
		repo,
		&mocks.MockOrderClient{
			GetOrderByIDFn: func(ctx context.Context, orderID uuid.UUID) (*orderclient.OrderResponse, error) {
				return &orderclient.OrderResponse{ID: orderID, UserID: uuid.New(), Status: "PAID"}, nil
			},
		},
		mustNewGenerator(t),
		mustNewVerifier(t),
		validator.New(),
		&mocks.MockAuthorizer{AuthorizeFn: func(role string, resource rbac.Resource, action rbac.Action) error { return nil }},
		mockClock,
		&mocks.MockNotifier{
			NotifyDeliveryLateFn: func(ctx context.Context, d *domain.Delivery) error { return nil },
		},
	)

	found, err := svc.GetByID(context.Background(), userID, delivery.ID, "customer")
	require.NoError(t, err)
	require.True(t, found.IsLate)
	require.Equal(t, domain.DeliveryStatusInTransit, found.Status)
}

// Scenario 10: Customer gets their own QR
func TestIntegration_CustomerGetsOwnQR(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	_, err := svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusInTransit)
	require.NoError(t, err)

	_, err = svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusReadyForPickup)
	require.NoError(t, err)

	qrResp, err := svc.GetQR(context.Background(), userID, delivery.ID, "customer")
	require.NoError(t, err)
	require.NotEmpty(t, qrResp.Token)
	require.False(t, qrResp.ExpiresAt.IsZero())
}

// Scenario 11: Customer tries to get QR for another's delivery — access denied
func TestIntegration_CustomerGetsOtherQR_AccessDenied(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	ownerID := uuid.New()
	otherID := uuid.New()
	delivery := createTestDelivery(t, svc, ownerID)

	_, err := svc.UpdateStatus(context.Background(), ownerID, delivery.ID, "employee", domain.DeliveryStatusInTransit)
	require.NoError(t, err)

	_, err = svc.UpdateStatus(context.Background(), ownerID, delivery.ID, "employee", domain.DeliveryStatusReadyForPickup)
	require.NoError(t, err)

	_, err = svc.GetQR(context.Background(), otherID, delivery.ID, "customer")
	require.ErrorIs(t, err, domain.ErrAccessDenied)
}

// Scenario 12: Modified QR token is rejected
func TestIntegration_ModifiedQR_Rejected(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	_, err := svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusInTransit)
	require.NoError(t, err)

	_, err = svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusReadyForPickup)
	require.NoError(t, err)

	qrResp, err := svc.GetQR(context.Background(), userID, delivery.ID, "customer")
	require.NoError(t, err)

	modifiedToken := qrResp.Token + "tampered"
	_, err = svc.VerifyQR(context.Background(), modifiedToken)
	require.Error(t, err)
}

// Scenario 13: Expired/revoked QR is rejected
func TestIntegration_RevokedQR_Rejected(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	_, err := svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusInTransit)
	require.NoError(t, err)

	_, err = svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusReadyForPickup)
	require.NoError(t, err)

	qrResp1, err := svc.GetQR(context.Background(), userID, delivery.ID, "customer")
	require.NoError(t, err)

	qrResp2, err := svc.GetQR(context.Background(), userID, delivery.ID, "customer")
	require.NoError(t, err)

	_, err = svc.VerifyQR(context.Background(), qrResp1.Token)
	require.ErrorIs(t, err, qr.ErrTokenRevoked)

	_, err = svc.VerifyQR(context.Background(), qrResp2.Token)
	require.NoError(t, err)
}

// Scenario 14: Full happy path — create → status changes → reschedule → get QR → verify
func TestIntegration_FullHappyPath(t *testing.T) {
	t.Parallel()

	svc, db := newIntegrationService(t)
	_ = db
	ctx := context.Background()

	userID := uuid.New()

	delivery, err := svc.Create(ctx, &domain.DeliveryCreateInput{
		UserID:                userID,
		OrderID:               uuid.New(),
		PickupAddress:         "789 Full Path St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		Role:                  "employee",
	})
	require.NoError(t, err)
	require.Equal(t, domain.DeliveryStatusPending, delivery.Status)

	delivery, err = svc.UpdateStatus(ctx, userID, delivery.ID, "employee", domain.DeliveryStatusInTransit)
	require.NoError(t, err)
	require.Equal(t, domain.DeliveryStatusInTransit, delivery.Status)

	delivery, err = svc.Reschedule(ctx, userID, delivery.ID, "employee", &domain.RescheduleInput{
		NewDate: time.Now().Add(72 * time.Hour),
		Reason:  "logistics",
	})
	require.NoError(t, err)
	require.False(t, delivery.EstimatedDeliveryDate.IsZero())

	delivery, err = svc.UpdateStatus(ctx, userID, delivery.ID, "employee", domain.DeliveryStatusReadyForPickup)
	require.NoError(t, err)
	require.Equal(t, domain.DeliveryStatusReadyForPickup, delivery.Status)

	qrResp, err := svc.GetQR(ctx, userID, delivery.ID, "customer")
	require.NoError(t, err)
	require.NotEmpty(t, qrResp.Token)

	verified, err := svc.VerifyQR(ctx, qrResp.Token)
	require.NoError(t, err)
	require.Equal(t, domain.DeliveryStatusDelivered, verified.Status)

	_, err = svc.GetQR(ctx, userID, delivery.ID, "customer")
	require.ErrorIs(t, err, domain.ErrQRNotAvailable)

	_, err = svc.VerifyQR(ctx, qrResp.Token)
	require.Error(t, err)
}

// Database persistence test
func TestIntegration_DatabasePersistence(t *testing.T) {
	t.Parallel()

	svc, db := newIntegrationService(t)
	ctx := context.Background()

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	var found domain.Delivery
	err := db.NewSelect().Model(&found).Where("id = ?", delivery.ID).Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, delivery.ID, found.ID)
	require.Equal(t, delivery.OrderID, found.OrderID)
	require.Equal(t, domain.DeliveryStatusPending, found.Status)
	require.Equal(t, "123 Test St", found.PickupAddress)
}

// Authorization test: employee can view any delivery
func TestIntegration_EmployeeCanViewAnyDelivery(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	otherEmployee := uuid.New()
	found, err := svc.GetByID(context.Background(), otherEmployee, delivery.ID, "employee")
	require.NoError(t, err)
	require.Equal(t, delivery.ID, found.ID)
}

// Error handling: duplicate delivery for same order
func TestIntegration_DuplicateDeliveryForOrder(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	orderID := uuid.New()

	input1 := &domain.DeliveryCreateInput{
		UserID:                userID,
		OrderID:               orderID,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		Role:                  "employee",
	}
	_, err := svc.Create(context.Background(), input1)
	require.NoError(t, err)

	input2 := &domain.DeliveryCreateInput{
		UserID:                userID,
		OrderID:               orderID,
		PickupAddress:         "456 Other St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		Role:                  "employee",
	}
	_, err = svc.Create(context.Background(), input2)
	require.Error(t, err)
}

// VerifyQR rejects token for wrong user
func TestIntegration_VerifyQR_WrongUser(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	ownerID := uuid.New()
	delivery := createTestDelivery(t, svc, ownerID)

	_, err := svc.UpdateStatus(context.Background(), ownerID, delivery.ID, "employee", domain.DeliveryStatusInTransit)
	require.NoError(t, err)

	_, err = svc.UpdateStatus(context.Background(), ownerID, delivery.ID, "employee", domain.DeliveryStatusReadyForPickup)
	require.NoError(t, err)

	impostorID := uuid.New()
	qrResp, err := svc.GetQR(context.Background(), impostorID, delivery.ID, "employee")
	require.NoError(t, err)

	_, err = svc.VerifyQR(context.Background(), qrResp.Token)
	require.ErrorIs(t, err, domain.ErrAccessDenied)
}

// Concurrent reschedule test
func TestIntegration_ConcurrentReschedules(t *testing.T) {
	t.Parallel()

	svc, _ := newIntegrationService(t)

	userID := uuid.New()
	delivery := createTestDelivery(t, svc, userID)

	_, err := svc.UpdateStatus(context.Background(), userID, delivery.ID, "employee", domain.DeliveryStatusInTransit)
	require.NoError(t, err)

	results := make(chan error, 2)

	go func() {
		_, err := svc.Reschedule(context.Background(), userID, delivery.ID, "employee", &domain.RescheduleInput{
			NewDate: time.Now().Add(72 * time.Hour),
			Reason:  "first reschedule",
		})
		results <- err
	}()

	go func() {
		_, err := svc.Reschedule(context.Background(), userID, delivery.ID, "employee", &domain.RescheduleInput{
			NewDate: time.Now().Add(96 * time.Hour),
			Reason:  "second reschedule",
		})
		results <- err
	}()

	err1 := <-results
	err2 := <-results

	require.NoError(t, err1)
	require.NoError(t, err2)
}

func newRepoFromService(t *testing.T, svc *DeliveryService) *deliveryrepo.DeliveryRepository {
	t.Helper()
	return svc.repo.(*deliveryrepo.DeliveryRepository)
}

func mustNewGenerator(t *testing.T) *qr.Generator {
	t.Helper()
	g, err := qr.NewGenerator("integration-test-secret-key-32-bytes!!", 24*time.Hour)
	require.NoError(t, err)
	return g
}

func mustNewVerifier(t *testing.T) *qr.Verifier {
	t.Helper()
	v, err := qr.NewVerifier("integration-test-secret-key-32-bytes!!")
	require.NoError(t, err)
	return v
}
