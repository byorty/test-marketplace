package repository

import (
	"github.com/byorty/test-marketplace/services/common/db"
	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
)

func toDBProduct(p *domain.Product) *db.Product {
	return &db.Product{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		Price:        p.Price,
		Category:     p.Category,
		Rating:       p.Rating,
		DeliveryDays: p.DeliveryDays,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func toDomainProduct(p *db.Product) *domain.Product {
	return &domain.Product{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		Price:        p.Price,
		Category:     p.Category,
		Rating:       p.Rating,
		DeliveryDays: p.DeliveryDays,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}