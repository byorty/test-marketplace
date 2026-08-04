package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ProductRepository struct {
	db domain.DB
	log *zap.Logger
}

func New(db domain.DB, log *zap.Logger) *ProductRepository {
	return &ProductRepository{
		db: db,
		log: log.Named("product-repository"),
	}
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	
	_, err := r.db.Exec(ctx, 
		`INSERT INTO products (id, name, description, category, price, delivery_days, rating, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID,
		p.Name,
		p.Description,
		p.Category,
		p.Price,
		p.DeliveryDays,
		p.Rating,
		p.CreatedAt,
		p.UpdatedAt,
	)

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

	err := r.db.QueryRow(
		ctx, `SELECT id, name, description, category, price, delivery_days, rating, created_at, updated_at
		FROM products WHERE id = $1`, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Category,
		&p.Price,
		&p.DeliveryDays,
		&p.Rating,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}

	if err != nil {

		r.log.Error(
			"get product failed",
			zap.Error(err),
			zap.String(
				"product_id",
				id.String(),
			),
		)

		return nil, fmt.Errorf("get product: %w", err)
	}

	return &p, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) (*domain.Product, error) {

	p.UpdatedAt = time.Now()

	var updated domain.Product

	err := r.db.QueryRow(ctx, 
		`UPDATE products 
		SET name = $1, description = $2, category = $3, price = $4, delivery_days = $5, rating = $6, updated_at = $7
		WHERE id = $8 RETURNING id, name, description, category, price, delivery_days, rating, created_at, updated_at`,
		p.Name,
		p.Description,
		p.Category,
		p.Price,
		p.DeliveryDays,
		p.Rating,
		p.UpdatedAt,
		p.ID,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Description,
		&updated.Category,
		&updated.Price,
		&updated.DeliveryDays,
		&updated.Rating,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {

		r.log.Error(
			"update product failed",
			zap.Error(err),
			zap.String(
				"product_id",
				p.ID.String(),
			),
		)

		return nil, fmt.Errorf("update product: %w", err)
	}

	return &updated, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {

	tag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)

	if err != nil {

		r.log.Error(
			"delete product failed",
			zap.Error(err),
			zap.String(
				"product_id",
				id.String(),
			),
		)

		return fmt.Errorf("delete product: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {

	args := make([]any, 0)
	where := make([]string, 0)

	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if filter.Name != "" {
		where = append(where,
			fmt.Sprintf("name ILIKE %s", addArg("%"+filter.Name+"%")),
		)
	}

	if filter.Category != "" {
		where = append(where,
			fmt.Sprintf("category = %s", addArg(filter.Category)),
		)
	}

	if filter.MinPrice != nil {
		where = append(where,
			fmt.Sprintf("price >= %s", addArg(*filter.MinPrice)),
		)
	}

	if filter.MaxPrice != nil {
		where = append(where,
			fmt.Sprintf("price <= %s", addArg(*filter.MaxPrice)),
		)
	}

	if filter.MinRating != nil {
		where = append(where,
			fmt.Sprintf("rating >= %s", addArg(*filter.MinRating)),
		)
	}

	if filter.MaxDeliveryDays != nil {
		where = append(where,
			fmt.Sprintf("delivery_days <= %s", addArg(*filter.MaxDeliveryDays)),
		)
	}

	baseQuery := " FROM products"

	if len(where) > 0 {
		baseQuery += " WHERE " + strings.Join(where, " AND ")
	}

	var total int64

	if err := r.db.QueryRow(
		ctx,
		"SELECT COUNT(*)"+baseQuery,
		args...,
	).Scan(&total); err != nil {

		r.log.Error(
			"count products failed",
			zap.Error(err),
		)

		return nil, err
	}


	orderBy := ""

	switch filter.SortBy {
	case domain.SortByPrice:
		orderBy = "price"
	case domain.SortByRating:
		orderBy = "rating"
	}

	if orderBy != "" {
		switch filter.Order {
		case domain.Desc:
			orderBy += " DESC"
		default:
			orderBy += " ASC"
		}
	}


	page := filter.Page
	if page <= 0 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize


	query := `
		SELECT 
			id,
			name,
			description,
			category,
			price,
			delivery_days,
			rating,
			created_at,
			updated_at
	` + baseQuery


	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}

	query += fmt.Sprintf(
		" LIMIT %d OFFSET %d",
		pageSize,
		offset,
	)


	rows, err := r.db.Query(ctx, query, args...)

	if err != nil {
		r.log.Error(
			"query products failed",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
		)

		return nil, err
	}

	defer rows.Close()


	items := make([]*domain.Product, 0)

	for rows.Next() {

		var p domain.Product

		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.Price,
			&p.DeliveryDays,
			&p.Rating,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {

			r.log.Error(
				"scan product failed",
				zap.Error(err),
			)

			return nil, err
		}

		items = append(items, &p)
	}


	if err := rows.Err(); err != nil {

		r.log.Error(
			"iterate products failed",
			zap.Error(err),
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