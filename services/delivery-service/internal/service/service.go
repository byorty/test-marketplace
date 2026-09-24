package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/client"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"github.com/byorty/test-marketplace/services/delivery-service/internal/qr"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeliveryService struct {
	repo        domain.DeliveryRepository
	log         *zap.Logger
	orderClient client.Client
	qrGen       *qr.Generator
	qrVer       *qr.Verifier
	validate    *validator.Validate
	authorizer  domain.Authorizer
	clock       domain.Clock
	notifier    domain.Notifier
}

func New(
	log *zap.Logger,
	repo domain.DeliveryRepository,
	orderClient client.Client,
	qrGen *qr.Generator,
	qrVer *qr.Verifier,
	validate *validator.Validate,
	authorizer domain.Authorizer,
	clock domain.Clock,
	notifier domain.Notifier,
) *DeliveryService {
	return &DeliveryService{
		repo:        repo,
		log:         log.Named("delivery-service"),
		orderClient: orderClient,
		qrGen:       qrGen,
		qrVer:       qrVer,
		validate:    validate,
		authorizer:  authorizer,
		clock:       clock,
		notifier:    notifier,
	}
}

func (s *DeliveryService) Create(ctx context.Context, input *domain.DeliveryCreateInput) (*domain.Delivery, error) {
	start := time.Now()

	if input == nil {
		s.log.Error("create delivery failed", zap.Error(ErrNilInput))
		return nil, ErrNilInput
	}

	if err := s.authorizer.Authorize(input.Role, rbac.ResourceDelivery, rbac.ActionCreate); err != nil {
		s.log.Error("create delivery failed", zap.Error(err), zap.String("role", input.Role))
		return nil, domain.ErrAccessDenied
	}

	if input.OrderID == uuid.Nil {
		s.log.Error("create delivery failed", zap.Error(ErrInvalidOrderID))
		return nil, ErrInvalidOrderID
	}

	if input.UserID == uuid.Nil {
		s.log.Error("create delivery failed", zap.Error(ErrInvalidUserID))
		return nil, ErrInvalidUserID
	}

	if err := s.validate.Var(input.PickupAddress, "required,min=1,max=500"); err != nil {
		s.log.Error("create delivery failed", zap.Error(err))
		return nil, ErrInvalidInput
	}

	if input.EstimatedDeliveryDate.IsZero() || input.EstimatedDeliveryDate.Before(s.clock.Now()) {
		s.log.Error("create delivery failed", zap.Error(ErrInvalidEstimatedDate))
		return nil, ErrInvalidEstimatedDate
	}

	order, err := s.orderClient.GetOrderByID(ctx, input.OrderID)
	if err != nil {
		s.log.Error("create delivery failed", zap.Error(err), zap.String("order_id", input.OrderID.String()))
		return nil, fmt.Errorf("get order: %w", err)
	}

	if order.Status != "PAID" {
		s.log.Error("create delivery failed", zap.Error(domain.ErrOrderNotPaid), zap.String("order_id", input.OrderID.String()))
		return nil, domain.ErrOrderNotPaid
	}

	existing, err := s.repo.GetByOrderID(ctx, input.OrderID)
	if err != nil && !errors.Is(err, domain.ErrDeliveryNotFound) {
		s.log.Error("create delivery failed", zap.Error(err), zap.String("order_id", input.OrderID.String()))
		return nil, fmt.Errorf("check existing delivery: %w", err)
	}
	if err == nil && existing != nil {
		if existing.Status != domain.DeliveryStatusCancelled {
			s.log.Error("create delivery failed", zap.Error(domain.ErrDeliveryAlreadyExists), zap.String("order_id", input.OrderID.String()))
			return nil, domain.ErrDeliveryAlreadyExists
		}
	}

	now := s.clock.Now()
	delivery := &domain.Delivery{
		ID:                    uuid.New(),
		OrderID:               input.OrderID,
		UserID:                order.UserID,
		Status:                domain.DeliveryStatusPending,
		PickupAddress:         input.PickupAddress,
		EstimatedDeliveryDate: input.EstimatedDeliveryDate,
		IsLate:                false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repo.Create(ctx, delivery); err != nil {
		s.log.Error("create delivery failed", zap.Error(err), zap.String("order_id", input.OrderID.String()))
		return nil, fmt.Errorf("create delivery: %w", err)
	}

	s.log.Info("delivery created", zap.String("delivery_id", delivery.ID.String()), zap.String("order_id", input.OrderID.String()), zap.Duration("duration", time.Since(start)))

	return delivery, nil
}

func (s *DeliveryService) GetByID(ctx context.Context, userID, id uuid.UUID, role string) (*domain.Delivery, error) {
	if id == uuid.Nil {
		s.log.Error("get delivery failed", zap.Error(ErrInvalidID))
		return nil, ErrInvalidID
	}

	delivery, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("get delivery failed", zap.Error(err), zap.String("delivery_id", id.String()))
		return nil, fmt.Errorf("get delivery: %w", err)
	}

	if role == string(rbac.RoleCustomer) && delivery.UserID != userID {
		s.log.Error("get delivery failed", zap.Error(domain.ErrAccessDenied), zap.String("delivery_id", id.String()), zap.String("user_id", userID.String()))
		return nil, domain.ErrAccessDenied
	}

	reschedules, err := s.repo.ListReschedules(ctx, delivery.ID)
	if err != nil {
		s.log.Error("get delivery reschedules failed", zap.Error(err), zap.String("delivery_id", id.String()))
		return nil, fmt.Errorf("get reschedules: %w", err)
	}
	delivery.Reschedules = make([]*domain.DeliveryReschedule, len(reschedules))
	for i := range reschedules {
		delivery.Reschedules[i] = &reschedules[i]
	}

	wasLate := delivery.IsLate
	delivery.IsLate = domain.IsOverdue(delivery, s.clock.Now())

	if delivery.IsLate && !wasLate {
		if err := s.notifier.NotifyDeliveryLate(ctx, delivery); err != nil {
			s.log.Error("notify delivery late failed", zap.Error(err), zap.String("delivery_id", id.String()))
		}
	}

	return delivery, nil
}

func (s *DeliveryService) List(ctx context.Context, userID uuid.UUID, role string, filter domain.DeliveryListFilter) (*domain.DeliveryList, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	if role == string(rbac.RoleCustomer) {
		filter.UserID = &userID
	}

	result, err := s.repo.List(ctx, filter)
	if err != nil {
		s.log.Error("list deliveries failed", zap.Error(err))
		return nil, fmt.Errorf("list deliveries: %w", err)
	}

	for i := range result.Items {
		result.Items[i].IsLate = domain.IsOverdue(result.Items[i], s.clock.Now())
	}

	return result, nil
}

func (s *DeliveryService) Reschedule(ctx context.Context, userID, deliveryID uuid.UUID, role string, input *domain.RescheduleInput) (*domain.Delivery, error) {
	start := time.Now()

	if err := s.authorizer.Authorize(role, rbac.ResourceDelivery, rbac.ActionReschedule); err != nil {
		s.log.Error("reschedule delivery failed", zap.Error(domain.ErrAccessDenied), zap.String("role", role))
		return nil, domain.ErrAccessDenied
	}

	if deliveryID == uuid.Nil {
		s.log.Error("reschedule delivery failed", zap.Error(ErrInvalidID))
		return nil, ErrInvalidID
	}

	if input == nil {
		s.log.Error("reschedule delivery failed", zap.Error(ErrNilInput))
		return nil, ErrNilInput
	}

	if input.NewDate.IsZero() || input.NewDate.Before(s.clock.Now()) {
		s.log.Error("reschedule delivery failed", zap.Error(ErrInvalidNewDate))
		return nil, ErrInvalidNewDate
	}

	delivery, err := s.repo.GetByID(ctx, deliveryID)
	if err != nil {
		s.log.Error("reschedule delivery failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, fmt.Errorf("get delivery: %w", err)
	}

	if !domain.CanReschedule(delivery.Status) {
		s.log.Error("reschedule delivery failed", zap.Error(domain.ErrRescheduleOnTerminal), zap.String("delivery_id", deliveryID.String()), zap.String("status", string(delivery.Status)))
		return nil, domain.ErrRescheduleOnTerminal
	}

	reschedule := &domain.DeliveryReschedule{
		ID:           uuid.New(),
		DeliveryID:   delivery.ID,
		PreviousDate: delivery.EstimatedDeliveryDate,
		NewDate:      input.NewDate,
		Reason:       input.Reason,
		CreatedAt:    s.clock.Now(),
	}

	delivery.EstimatedDeliveryDate = input.NewDate
	delivery.IsLate = s.clock.Now().After(input.NewDate) && !delivery.Status.IsTerminal()
	delivery.UpdatedAt = s.clock.Now()

	updated, err := s.repo.RescheduleTx(ctx, delivery, reschedule)
	if err != nil {
		s.log.Error("reschedule delivery failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, fmt.Errorf("reschedule delivery: %w", err)
	}

	reschedules, err := s.repo.ListReschedules(ctx, updated.ID)
	if err != nil {
		s.log.Error("reschedule delivery: load reschedules failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, fmt.Errorf("load reschedules: %w", err)
	}
	updated.Reschedules = make([]*domain.DeliveryReschedule, len(reschedules))
	for i := range reschedules {
		updated.Reschedules[i] = &reschedules[i]
	}

	s.log.Info("delivery rescheduled", zap.String("delivery_id", deliveryID.String()), zap.Duration("duration", time.Since(start)))

	return updated, nil
}

func (s *DeliveryService) UpdateStatus(ctx context.Context, userID, deliveryID uuid.UUID, role string, newStatus domain.DeliveryStatus) (*domain.Delivery, error) {
	start := time.Now()

	if err := s.authorizer.Authorize(role, rbac.ResourceDelivery, rbac.ActionUpdateStatus); err != nil {
		s.log.Error("update delivery status failed", zap.Error(domain.ErrAccessDenied), zap.String("role", role))
		return nil, domain.ErrAccessDenied
	}

	if deliveryID == uuid.Nil {
		s.log.Error("update delivery status failed", zap.Error(ErrInvalidID))
		return nil, ErrInvalidID
	}

	if !newStatus.IsValid() {
		s.log.Error("update delivery status failed", zap.Error(ErrInvalidStatus))
		return nil, ErrInvalidStatus
	}

	delivery, err := s.repo.GetByID(ctx, deliveryID)
	if err != nil {
		s.log.Error("update delivery status failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, fmt.Errorf("get delivery: %w", err)
	}

	if delivery.Status.IsTerminal() {
		s.log.Error("update delivery status failed", zap.Error(domain.ErrDeliveryAlreadyDelivered), zap.String("delivery_id", deliveryID.String()), zap.String("current_status", string(delivery.Status)))
		if delivery.Status == domain.DeliveryStatusDelivered {
			return nil, domain.ErrDeliveryAlreadyDelivered
		}
		return nil, domain.ErrDeliveryAlreadyCancelled
	}

	if !domain.CanTransition(delivery.Status, newStatus) {
		s.log.Error("update delivery status failed", zap.Error(domain.ErrInvalidStatusTransition), zap.String("delivery_id", deliveryID.String()), zap.String("from", string(delivery.Status)), zap.String("to", string(newStatus)))
		return nil, domain.ErrInvalidStatusTransition
	}

	oldStatus := delivery.Status
	delivery.Status = newStatus
	delivery.UpdatedAt = s.clock.Now()

	if newStatus == domain.DeliveryStatusDelivered || newStatus == domain.DeliveryStatusCancelled {
		delivery.QRNonce = uuid.NullUUID{}
	}

	updated, err := s.repo.Update(ctx, delivery)
	if err != nil {
		s.log.Error("update delivery status failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, fmt.Errorf("update delivery: %w", err)
	}

	if err := s.notifier.NotifyDeliveryStatusChanged(ctx, updated, oldStatus); err != nil {
		s.log.Error("notify delivery status changed failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
	}

	s.log.Info("delivery status updated", zap.String("delivery_id", deliveryID.String()), zap.String("status", string(newStatus)), zap.Duration("duration", time.Since(start)))

	return updated, nil
}

func (s *DeliveryService) GetQR(ctx context.Context, userID, deliveryID uuid.UUID, role string) (*qr.QRTokenResponse, error) {
	if deliveryID == uuid.Nil {
		s.log.Error("get QR failed", zap.Error(ErrInvalidID))
		return nil, ErrInvalidID
	}

	delivery, err := s.repo.GetByID(ctx, deliveryID)
	if err != nil {
		s.log.Error("get QR failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, fmt.Errorf("get delivery: %w", err)
	}

	if role == string(rbac.RoleCustomer) && delivery.UserID != userID {
		s.log.Error("get QR failed", zap.Error(domain.ErrAccessDenied), zap.String("delivery_id", deliveryID.String()), zap.String("user_id", userID.String()))
		return nil, domain.ErrAccessDenied
	}

	if delivery.Status != domain.DeliveryStatusReadyForPickup {
		s.log.Error("get QR failed", zap.Error(domain.ErrQRNotAvailable), zap.String("delivery_id", deliveryID.String()), zap.String("status", string(delivery.Status)))
		return nil, domain.ErrQRNotAvailable
	}

	nonce := uuid.New()
	delivery.QRNonce = uuid.NullUUID{UUID: nonce, Valid: true}
	delivery.UpdatedAt = s.clock.Now()

	updated, err := s.repo.Update(ctx, delivery)
	if err != nil {
		s.log.Error("get QR failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, fmt.Errorf("update delivery: %w", err)
	}

	response, err := s.qrGen.Generate(updated.ID, updated.OrderID, userID, nonce)
	if err != nil {
		s.log.Error("get QR failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, fmt.Errorf("generate QR: %w", err)
	}

	return response, nil
}

func (s *DeliveryService) VerifyQR(ctx context.Context, token string) (*domain.Delivery, error) {
	qrToken, err := s.qrVer.Verify(token)
	if err != nil {
		s.log.Error("verify QR failed", zap.Error(err))
		return nil, fmt.Errorf("verify QR token: %w", err)
	}

	delivery, err := s.repo.GetByID(ctx, qrToken.DeliveryID)
	if err != nil {
		s.log.Error("verify QR failed", zap.Error(err), zap.String("delivery_id", qrToken.DeliveryID.String()))
		return nil, fmt.Errorf("get delivery: %w", err)
	}

	if delivery.Status != domain.DeliveryStatusReadyForPickup {
		s.log.Error("verify QR failed", zap.Error(domain.ErrQRNotAvailable), zap.String("delivery_id", qrToken.DeliveryID.String()), zap.String("status", string(delivery.Status)))
		return nil, domain.ErrQRNotAvailable
	}

	if qrToken.UserID != delivery.UserID {
		s.log.Error("verify QR failed", zap.Error(domain.ErrAccessDenied), zap.String("delivery_id", qrToken.DeliveryID.String()), zap.String("token_user_id", qrToken.UserID.String()), zap.String("delivery_user_id", delivery.UserID.String()))
		return nil, domain.ErrAccessDenied
	}

	if !delivery.QRNonce.Valid || qrToken.Nonce != delivery.QRNonce.UUID {
		s.log.Error("verify QR failed", zap.Error(qr.ErrTokenRevoked), zap.String("delivery_id", qrToken.DeliveryID.String()))
		return nil, qr.ErrTokenRevoked
	}

	updated, err := s.repo.UpdateStatusConditional(
		ctx,
		delivery.ID,
		domain.DeliveryStatusReadyForPickup,
		domain.DeliveryStatusDelivered,
		map[string]interface{}{
			"status":     string(domain.DeliveryStatusDelivered),
			"qr_nonce":   nil,
			"updated_at": s.clock.Now(),
		},
	)
	if err != nil {
		s.log.Error("verify QR failed", zap.Error(err), zap.String("delivery_id", qrToken.DeliveryID.String()))
		return nil, fmt.Errorf("update delivery status: %w", err)
	}

	if err := s.orderClient.UpdateOrderStatus(ctx, updated.OrderID, "DELIVERED"); err != nil {
		s.log.Error("verify QR: update order status failed", zap.Error(err), zap.String("order_id", updated.OrderID.String()), zap.String("delivery_id", qrToken.DeliveryID.String()))
	}

	s.log.Info("QR verified, delivery delivered", zap.String("delivery_id", qrToken.DeliveryID.String()))

	return updated, nil
}
