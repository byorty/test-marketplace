package repository

import (
	"context"
	"testing"
	"time"

	testtools "github.com/byorty/test-marketplace/services/common/test-tools"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

func newTestRepository(t *testing.T) (*DeliveryRepository, *bun.DB) {
	t.Helper()

	database := testtools.NewTestDB(t)

	repo := New(database, zap.NewNop())

	return repo, database
}

func TestDeliveryRepository_Create(t *testing.T) {
	t.Parallel()

	repo, db := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	var found domain.Delivery
	err = db.NewSelect().Model(&found).Where("id = ?", delivery.ID).Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, delivery.ID, found.ID)
	require.Equal(t, delivery.OrderID, found.OrderID)
	require.Equal(t, domain.DeliveryStatusPending, found.Status)
	require.Equal(t, delivery.PickupAddress, found.PickupAddress)
}

func TestDeliveryRepository_GetByID(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusInTransit,
		PickupAddress:         "456 Oak Ave",
		EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, delivery.ID)
	require.NoError(t, err)
	require.Equal(t, delivery.ID, found.ID)
	require.Equal(t, delivery.OrderID, found.OrderID)
	require.Equal(t, domain.DeliveryStatusInTransit, found.Status)
}

func TestDeliveryRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, uuid.New())
	require.ErrorIs(t, err, domain.ErrDeliveryNotFound)
	require.Nil(t, found)
}

func TestDeliveryRepository_GetByID_WithReschedules(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusInTransit,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	reschedule1 := &domain.DeliveryReschedule{
		ID:           uuid.New(),
		DeliveryID:   delivery.ID,
		PreviousDate: delivery.EstimatedDeliveryDate,
		NewDate:      time.Now().Add(48 * time.Hour),
		Reason:       "weather delay",
		CreatedAt:    time.Now(),
	}
	reschedule2 := &domain.DeliveryReschedule{
		ID:           uuid.New(),
		DeliveryID:   delivery.ID,
		PreviousDate: reschedule1.NewDate,
		NewDate:      time.Now().Add(72 * time.Hour),
		Reason:       "traffic",
		CreatedAt:    time.Now(),
	}

	err = repo.CreateReschedule(ctx, reschedule1)
	require.NoError(t, err)
	err = repo.CreateReschedule(ctx, reschedule2)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, delivery.ID)
	require.NoError(t, err)
	require.Equal(t, delivery.ID, found.ID)
	require.Equal(t, delivery.UserID, found.UserID)
	require.Equal(t, delivery.OrderID, found.OrderID)

	reschedules, err := repo.ListReschedules(ctx, delivery.ID)
	require.NoError(t, err)
	require.Len(t, reschedules, 2)
	require.Equal(t, reschedule1.ID, reschedules[0].ID)
	require.Equal(t, reschedule2.ID, reschedules[1].ID)
}

func TestDeliveryRepository_GetByID_UserIDFilter(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	userA := uuid.New()
	userB := uuid.New()

	deliveryA := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                userA,
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	deliveryB := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                userB,
		Status:                domain.DeliveryStatusInTransit,
		PickupAddress:         "456 Oak Ave",
		EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, deliveryA)
	require.NoError(t, err)
	err = repo.Create(ctx, deliveryB)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, deliveryA.ID)
	require.NoError(t, err)
	require.Equal(t, userA, found.UserID)

	found, err = repo.GetByID(ctx, deliveryB.ID)
	require.NoError(t, err)
	require.Equal(t, userB, found.UserID)
}

func TestDeliveryRepository_GetByOrderID(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	orderID := uuid.New()
	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               orderID,
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "789 Elm St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	found, err := repo.GetByOrderID(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, delivery.ID, found.ID)
	require.Equal(t, orderID, found.OrderID)
}

func TestDeliveryRepository_GetByOrderID_NotFound(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	found, err := repo.GetByOrderID(ctx, uuid.New())
	require.ErrorIs(t, err, domain.ErrDeliveryNotFound)
	require.Nil(t, found)
}

func TestDeliveryRepository_Update(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	delivery.Status = domain.DeliveryStatusInTransit
	delivery.UpdatedAt = time.Now()

	updated, err := repo.Update(ctx, delivery)
	require.NoError(t, err)
	require.Equal(t, domain.DeliveryStatusInTransit, updated.Status)
}

func TestDeliveryRepository_Update_StatusTransitions(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	t.Run("pending -> in_transit", func(t *testing.T) {
		delivery.Status = domain.DeliveryStatusInTransit
		delivery.UpdatedAt = time.Now()
		updated, err := repo.Update(ctx, delivery)
		require.NoError(t, err)
		require.Equal(t, domain.DeliveryStatusInTransit, updated.Status)
		delivery = updated
	})

	t.Run("in_transit -> ready_for_pickup", func(t *testing.T) {
		delivery.Status = domain.DeliveryStatusReadyForPickup
		delivery.UpdatedAt = time.Now()
		updated, err := repo.Update(ctx, delivery)
		require.NoError(t, err)
		require.Equal(t, domain.DeliveryStatusReadyForPickup, updated.Status)
		delivery = updated
	})

	t.Run("ready_for_pickup -> delivered", func(t *testing.T) {
		delivery.Status = domain.DeliveryStatusDelivered
		delivery.UpdatedAt = time.Now()
		updated, err := repo.Update(ctx, delivery)
		require.NoError(t, err)
		require.Equal(t, domain.DeliveryStatusDelivered, updated.Status)
		require.Equal(t, delivery.ID, updated.ID)
		require.Equal(t, delivery.OrderID, updated.OrderID)
	})
}

func TestDeliveryRepository_Update_CancelFromAnyNonTerminal(t *testing.T) {
	t.Parallel()

	for _, initialStatus := range []domain.DeliveryStatus{
		domain.DeliveryStatusPending,
		domain.DeliveryStatusInTransit,
		domain.DeliveryStatusReadyForPickup,
	} {
		t.Run("cancel from "+string(initialStatus), func(t *testing.T) {
			t.Parallel()

			repo, _ := newTestRepository(t)
			ctx := context.Background()

			delivery := &domain.Delivery{
				ID:                    uuid.New(),
				OrderID:               uuid.New(),
				UserID:                uuid.New(),
				Status:                initialStatus,
				PickupAddress:         "123 Main St",
				EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
				IsLate:                false,
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			}

			err := repo.Create(ctx, delivery)
			require.NoError(t, err)

			delivery.Status = domain.DeliveryStatusCancelled
			delivery.UpdatedAt = time.Now()

			updated, err := repo.Update(ctx, delivery)
			require.NoError(t, err)
			require.Equal(t, domain.DeliveryStatusCancelled, updated.Status)
		})
	}
}

func TestDeliveryRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	_, err := repo.Update(ctx, delivery)
	require.ErrorIs(t, err, domain.ErrDeliveryNotFound)
}

func TestDeliveryRepository_List(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	userID := uuid.New()

	for i := 0; i < 5; i++ {
		delivery := &domain.Delivery{
			ID:                    uuid.New(),
			OrderID:               uuid.New(),
			UserID:                userID,
			Status:                domain.DeliveryStatusPending,
			PickupAddress:         "123 Main St",
			EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
			IsLate:                false,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}
		err := repo.Create(ctx, delivery)
		require.NoError(t, err)
	}

	for i := 0; i < 3; i++ {
		delivery := &domain.Delivery{
			ID:                    uuid.New(),
			OrderID:               uuid.New(),
			UserID:                uuid.New(),
			Status:                domain.DeliveryStatusInTransit,
			PickupAddress:         "456 Oak Ave",
			EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
			IsLate:                false,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}
		err := repo.Create(ctx, delivery)
		require.NoError(t, err)
	}

	t.Run("filter_by_status", func(t *testing.T) {
		status := domain.DeliveryStatusPending
		filter := domain.DeliveryListFilter{
			Status:   &status,
			Page:     1,
			PageSize: 10,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		require.Equal(t, int64(5), result.Total)
		require.Len(t, result.Items, 5)
	})

	t.Run("filter_by_user_id", func(t *testing.T) {
		filter := domain.DeliveryListFilter{
			UserID:   &userID,
			Page:     1,
			PageSize: 10,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		require.Equal(t, int64(5), result.Total)
		require.Len(t, result.Items, 5)
	})

	t.Run("filter_by_user_id_and_status", func(t *testing.T) {
		status := domain.DeliveryStatusPending
		filter := domain.DeliveryListFilter{
			UserID:   &userID,
			Status:   &status,
			Page:     1,
			PageSize: 10,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		require.Equal(t, int64(5), result.Total)
		require.Len(t, result.Items, 5)
	})

	t.Run("list_all", func(t *testing.T) {
		filter := domain.DeliveryListFilter{
			Page:     1,
			PageSize: 10,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		require.Equal(t, int64(8), result.Total)
		require.Len(t, result.Items, 8)
	})
}

func TestDeliveryRepository_CreateReschedule(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	reschedule := &domain.DeliveryReschedule{
		ID:           uuid.New(),
		DeliveryID:   delivery.ID,
		PreviousDate: delivery.EstimatedDeliveryDate,
		NewDate:      time.Now().Add(72 * time.Hour),
		Reason:       "weather delay",
		CreatedAt:    time.Now(),
	}

	err = repo.CreateReschedule(ctx, reschedule)
	require.NoError(t, err)

	reschedules, err := repo.ListReschedules(ctx, delivery.ID)
	require.NoError(t, err)
	require.Len(t, reschedules, 1)
	require.Equal(t, reschedule.ID, reschedules[0].ID)
	require.Equal(t, delivery.ID, reschedules[0].DeliveryID)
}

func TestDeliveryRepository_ListReschedules_Empty(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	reschedules, err := repo.ListReschedules(ctx, delivery.ID)
	require.NoError(t, err)
	require.Empty(t, reschedules)
}

func TestDeliveryRepository_CreateReschedule_RepeatedReschedule(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusInTransit,
		PickupAddress:         "123 Main St",
		EstimatedDeliveryDate: time.Now().Add(48 * time.Hour),
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	reschedule1 := &domain.DeliveryReschedule{
		ID:           uuid.New(),
		DeliveryID:   delivery.ID,
		PreviousDate: delivery.EstimatedDeliveryDate,
		NewDate:      time.Now().Add(72 * time.Hour),
		Reason:       "weather delay",
		CreatedAt:    time.Now(),
	}

	err = repo.CreateReschedule(ctx, reschedule1)
	require.NoError(t, err)

	newEstimatedDate := time.Now().Add(72 * time.Hour)
	delivery.EstimatedDeliveryDate = newEstimatedDate
	delivery.UpdatedAt = time.Now()
	_, err = repo.Update(ctx, delivery)
	require.NoError(t, err)

	reschedule2 := &domain.DeliveryReschedule{
		ID:           uuid.New(),
		DeliveryID:   delivery.ID,
		PreviousDate: newEstimatedDate,
		NewDate:      time.Now().Add(96 * time.Hour),
		Reason:       "traffic delay",
		CreatedAt:    time.Now(),
	}

	err = repo.CreateReschedule(ctx, reschedule2)
	require.NoError(t, err)

	reschedules, err := repo.ListReschedules(ctx, delivery.ID)
	require.NoError(t, err)
	require.Len(t, reschedules, 2)

	require.Equal(t, reschedule1.ID, reschedules[0].ID)
	require.Equal(t, reschedule1.PreviousDate.UTC().Truncate(time.Millisecond), reschedules[0].PreviousDate.UTC().Truncate(time.Millisecond))
	require.Equal(t, reschedule1.NewDate.UTC().Truncate(time.Millisecond), reschedules[0].NewDate.UTC().Truncate(time.Millisecond))
	require.Equal(t, "weather delay", reschedules[0].Reason)

	require.Equal(t, reschedule2.ID, reschedules[1].ID)
	require.Equal(t, newEstimatedDate.UTC().Truncate(time.Millisecond), reschedules[1].PreviousDate.UTC().Truncate(time.Millisecond))
	require.Equal(t, reschedule2.NewDate.UTC().Truncate(time.Millisecond), reschedules[1].NewDate.UTC().Truncate(time.Millisecond))
	require.Equal(t, "traffic delay", reschedules[1].Reason)
}

func TestDeliveryRepository_Reschedule_AuditFields(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepository(t)
	ctx := context.Background()

	originalDate := time.Now().Add(48 * time.Hour)

	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               uuid.New(),
		UserID:                uuid.New(),
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         "789 Elm St",
		EstimatedDeliveryDate: originalDate,
		IsLate:                false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	err := repo.Create(ctx, delivery)
	require.NoError(t, err)

	reschedule := &domain.DeliveryReschedule{
		ID:           uuid.New(),
		DeliveryID:   delivery.ID,
		PreviousDate: originalDate,
		NewDate:      time.Now().Add(72 * time.Hour),
		Reason:       "weather",
		CreatedAt:    time.Now(),
	}

	err = repo.CreateReschedule(ctx, reschedule)
	require.NoError(t, err)

	reschedules, err := repo.ListReschedules(ctx, delivery.ID)
	require.NoError(t, err)
	require.Len(t, reschedules, 1)

	r := reschedules[0]
	require.Equal(t, delivery.ID, r.DeliveryID)
	require.Equal(t, originalDate.UTC().Truncate(time.Millisecond), r.PreviousDate.UTC().Truncate(time.Millisecond))
	require.False(t, r.NewDate.IsZero())
	require.Equal(t, "weather", r.Reason)
	require.False(t, r.CreatedAt.IsZero())
}
