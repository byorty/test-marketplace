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
	model := toModel(p)

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

func (r *Repository) Update(ctx context.Context, p *domain.Product) error {
	const op = "Repository.Update"
	p.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&ProductModel{}).Where("id = ?", p.ID).Updates(toModel(p))

	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "Repository.Delete"
	result := r.db.WithContext(ctx).Delete(&ProductModel{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}



func (r *Repository) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
	const op = "Repository.List"

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

	if filter.MaxDeliveryDays != nil {
		query = query.Where("delivery_days <= ?", *filter.MaxDeliveryDays)
	}

	orderExpr := resolveOrder(filter.SortBy, filter.Order) 
	if orderExpr != "" {
		query = query.Order(orderExpr)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("%s: count: %w", op, err)
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	
	offset := (page-1) * pageSize

	var models []ProductModel

	if err := query.Limit(pageSize).Offset(offset).Find(&models).Error; 
	err != nil {
		return nil, fmt.Errorf("%s: find: %w", op, err)
	}

	items := make([]domain.Product, 0, len(models))
	for _, m := range models {
		items = append(items, *toDomain(m))
	}

	return &domain.ProductList{
		Items: items,
		Total: total,
		Page: page,
		PageSize: pageSize,
	}, nil
}

func resolveOrder(sortBy domain.SortBy, order domain.SortOrder) string {
	var column string

	switch sortBy {
	case domain.SortByPrice:
		column = "price"
	case domain.SortByRating:
		column = "rating"
	default:
		return ""
	}

	dir := "ASC"

	switch order {
	case domain.Desc:
		dir = "DESC"
	case domain.Asc:
		dir = "ASC"
	}

	return column + " " + dir
}