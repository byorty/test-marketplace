package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type ProductRepository struct {
	db *bun.DB
	log *zap.Logger
}

func New(db *bun.DB, log *zap.Logger) *ProductRepository {
	return &ProductRepository{
		db: db,
		log: log.Named("product-repository"),
	}
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	
	_, err := r.db.NewInsert().Model(p).Exec(ctx)

	if err != nil {
		r.log.Error(
			"create product failed",
			zap.Error(err),
			zap.String("product_id", p.ID.String()),
		)
		return err
	}

	r.log.Info(
		"product created",
		zap.String("product_id", p.ID.String()),
	)

	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {

	var p domain.Product

	err := r.db.NewSelect().Model(&p).Where("id = ?", id).Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}

	if err != nil {
		r.log.Error(
			"get product failed",
			zap.Error(err),
			zap.String("product_id", id.String()),
		)
		return nil, err
	}

	return &p, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) (*domain.Product, error) {

	p.UpdatedAt = time.Now()

	updated := new(domain.Product)

	err := r.db.NewUpdate().
		Model(p).
		Where("id = ?", p.ID).
		Returning("*").
		Scan(ctx, updated)

	if err != nil {
		r.log.Error(
			"update product failed",
			zap.Error(err),
			zap.String("product_id", p.ID.String()),
		)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return updated, nil
}	

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {

	res, err := r.db.NewDelete().Model((*domain.Product)(nil)).Where("id = ?", id).Exec(ctx)

	if err != nil {
		r.log.Error(
			"delete product failed",
			zap.Error(err),
			zap.String("product_id", id.String(),),
		)

		return fmt.Errorf("delete product: %w", err)
	}

	rows, err := res.RowsAffected()

	if rows == 0 {
		return domain.ErrProductNotFound
	}

	r.log.Info(
		"product deleted",
		zap.String("product_id", id.String()),
	)

	return nil
}

func (r *ProductRepository) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	query := r.db.NewSelect().Model((*domain.Product)(nil))

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}

	if filter.MinRating != nil {
		query = query.Where("rating >= ?", *filter.MinRating)
	}

	if filter.MaxDeliveryDays != nil {
		query = query.Where(
			"delivery_days <= ?",
			*filter.MaxDeliveryDays,
		)
	}

	var total int64

	countQuery := query.Clone()

	count, err := countQuery.Count(ctx)
	if err != nil {
		r.log.Error(
			"count products failed",
			zap.Error(err),
		)

		return nil, err
	}

	total = int64(count)

	switch filter.SortBy {
	case domain.SortByPrice:
		if filter.Order == domain.Desc {
			query = query.OrderExpr("price DESC")
		} else {
			query = query.OrderExpr("price ASC")
		}

	case domain.SortByRating:
		if filter.Order == domain.Desc {
			query = query.OrderExpr("rating DESC")
		} else {
			query = query.OrderExpr("rating ASC")
		}
	}

	offset := (page - 1) * pageSize

	items := make([]*domain.Product, 0)

	err = query.
		Limit(pageSize).
		Offset(offset).
		Scan(ctx, &items)

	if err != nil {
		r.log.Error(
			"query products failed",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
		)

		return nil, err
	}

	r.log.Info(
		"products listed",
		zap.Int("count", len(items)),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	return &domain.ProductList{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}