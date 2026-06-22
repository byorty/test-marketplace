package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, p *domain.Product) error {
	const op = "Repository.Create"
	model := fromDomain(p)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	const op = "Repository.GetByID"
	var model ProductModel

	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return toDomain(model), nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "Repository.Delete"
	result := r.db.WithContext(ctx).Delete(&ProductModel{}, "id = ?", id)

	if result.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, update domain.UpdateProduct) error {
	const op = "Repository.Update"

	updates := make(map[string]any)

	if update.Name != nil {
		updates["name"] = *update.Name
	}

	if update.Description != nil {
		updates["description"] = *update.Description
	}

	if update.Category != nil {
		updates["category"] = *update.Category
	}

	if update.Price != nil {
		updates["price"] = *update.Price
	}

	if update.DeliveryDays != nil {
		updates["delivery_days"] = *update.DeliveryDays
	}

	if len(updates) == 0 {
		return domain.ErrEmptyUpdate
	}

	updates["updated_at"] = time.Now()

	result := r.db.WithContext(ctx).Model(&ProductModel{}).Where("id = ?", id).Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
	query := r.db.WithContext(ctx).Model(&ProductModel{}) 

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%" + filter.Name + "%")
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

	if filter.DeliveryDays != nil {
		query = query.Where("delivery_days <= ?", *filter.DeliveryDays)
	}

	var orderClause string

	switch filter.SortBy {
	case domain.SortByPrice:
		orderClause = "price"

	case domain.SortByRating:
		orderClause = "rating"
	}

	if orderClause != "" {
		switch filter.Order {
		case domain.Desc:
			orderClause += " DESC"

		default:
			orderClause += " ASC"
		}

		query = query.Order(orderClause)
	}

	var total int64

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("Repository.List.Count: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize

	query = query.Limit(filter.PageSize).Offset(offset)

	var models []ProductModel

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("Repository.List.Find: %w", err)
	}

	items := make([]domain.Product, len(models))

	for i, m := range models {
		items[i] = *toDomain(m)
	}

	return &domain.ProductList{
		Items: items,
		Total: total,
		Page: filter.Page,
		PageSize: filter.PageSize,
	}, nil
}