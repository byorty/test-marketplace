package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/byorty/test-marketplace/services/delivery-service/internal/domain"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type DeliveryRepository struct {
	db  *bun.DB
	log *zap.Logger
}

func New(db *bun.DB, log *zap.Logger) *DeliveryRepository {
	return &DeliveryRepository{
		db:  db,
		log: log.Named("delivery-repository"),
	}
}

func (r *DeliveryRepository) Create(ctx context.Context, delivery *domain.Delivery) error {
	_, err := r.db.NewInsert().Model(delivery).Exec(ctx)
	if err != nil {
		r.log.Error("create delivery failed", zap.Error(err), zap.String("order_id", delivery.OrderID.String()))
		return err
	}

	r.log.Info("delivery created", zap.String("delivery_id", delivery.ID.String()))
	return nil
}

func (r *DeliveryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	var delivery domain.Delivery

	err := r.db.NewSelect().Model(&delivery).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrDeliveryNotFound
	}
	if err != nil {
		r.log.Error("get delivery failed", zap.Error(err), zap.String("delivery_id", id.String()))
		return nil, err
	}

	return &delivery, nil
}

func (r *DeliveryRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Delivery, error) {
	var delivery domain.Delivery

	err := r.db.NewSelect().Model(&delivery).Where("order_id = ?", orderID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrDeliveryNotFound
	}
	if err != nil {
		r.log.Error("get delivery by order_id failed", zap.Error(err), zap.String("order_id", orderID.String()))
		return nil, err
	}

	return &delivery, nil
}

func (r *DeliveryRepository) List(ctx context.Context, filter domain.DeliveryListFilter) (*domain.DeliveryList, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	query := r.db.NewSelect().Model((*domain.Delivery)(nil))

	if filter.Status != nil {
		query = query.Where("status = ?", string(*filter.Status))
	}
	if filter.IsLate != nil {
		query = query.Where("is_late = ?", *filter.IsLate)
	}
	if filter.OrderID != nil {
		query = query.Where("order_id = ?", filter.OrderID.String())
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", filter.UserID.String())
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.log.Error("count deliveries failed", zap.Error(err))
		return nil, err
	}

	offset := (page - 1) * pageSize
	items := make([]*domain.Delivery, 0)

	err = query.
		OrderExpr("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(ctx, &items)
	if err != nil {
		r.log.Error("list deliveries failed", zap.Error(err))
		return nil, err
	}

	return &domain.DeliveryList{
		Items:    items,
		Total:    int64(count),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *DeliveryRepository) Update(ctx context.Context, delivery *domain.Delivery) (*domain.Delivery, error) {
	updated := new(domain.Delivery)

	err := r.db.NewUpdate().
		Model(delivery).
		Where("id = ?", delivery.ID).
		Returning("*").
		Scan(ctx, updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDeliveryNotFound
		}
		r.log.Error("update delivery failed", zap.Error(err), zap.String("delivery_id", delivery.ID.String()))
		return nil, err
	}

	return updated, nil
}

func (r *DeliveryRepository) CreateReschedule(ctx context.Context, reschedule *domain.DeliveryReschedule) error {
	_, err := r.db.NewInsert().Model(reschedule).Exec(ctx)
	if err != nil {
		r.log.Error("create reschedule failed", zap.Error(err), zap.String("delivery_id", reschedule.DeliveryID.String()))
		return err
	}

	r.log.Info("reschedule created", zap.String("reschedule_id", reschedule.ID.String()))
	return nil
}

func (r *DeliveryRepository) ListReschedules(ctx context.Context, deliveryID uuid.UUID) ([]domain.DeliveryReschedule, error) {
	var reschedules []domain.DeliveryReschedule

	err := r.db.NewSelect().
		Model(&reschedules).
		Where("delivery_id = ?", deliveryID).
		OrderExpr("created_at ASC").
		Scan(ctx)
	if err != nil {
		r.log.Error("list reschedules failed", zap.Error(err), zap.String("delivery_id", deliveryID.String()))
		return nil, err
	}

	return reschedules, nil
}

func (r *DeliveryRepository) UpdateStatusConditional(ctx context.Context, id uuid.UUID, fromStatus domain.DeliveryStatus, toStatus domain.DeliveryStatus, updates map[string]interface{}) (*domain.Delivery, error) {
	updated := new(domain.Delivery)

	q := r.db.NewUpdate().
		Model((*domain.Delivery)(nil)).
		Where("id = ?", id).
		Where("status = ?", string(fromStatus))

	for k, v := range updates {
		q = q.Set("? = ?", bun.Ident(k), v)
	}

	err := q.Returning("*").Scan(ctx, updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrQRNotAvailable
		}
		r.log.Error("conditional update delivery failed", zap.Error(err), zap.String("delivery_id", id.String()))
		return nil, err
	}

	return updated, nil
}

func (r *DeliveryRepository) RescheduleTx(ctx context.Context, delivery *domain.Delivery, reschedule *domain.DeliveryReschedule) (*domain.Delivery, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.log.Error("begin transaction failed", zap.Error(err))
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.NewInsert().Model(reschedule).Exec(ctx)
	if err != nil {
		r.log.Error("create reschedule in tx failed", zap.Error(err), zap.String("delivery_id", delivery.ID.String()))
		return nil, err
	}

	updated := new(domain.Delivery)
	err = tx.NewUpdate().
		Model(delivery).
		Where("id = ?", delivery.ID).
		Returning("*").
		Scan(ctx, updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDeliveryNotFound
		}
		r.log.Error("update delivery in tx failed", zap.Error(err), zap.String("delivery_id", delivery.ID.String()))
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		r.log.Error("commit transaction failed", zap.Error(err))
		return nil, err
	}

	return updated, nil
}
