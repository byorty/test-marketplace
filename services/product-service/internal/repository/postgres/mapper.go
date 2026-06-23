package postgres

import domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"

func toDomain(m ProductModel) *domain.Product{
	return &domain.Product{
		ID: m.ID,
		Name: m.Name,
		Description: m.Description,
		Price: m.Price,
		Category: m.Category,
		Rating: m.Rating,
		DeliveryDays: m.DeliveryDays,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toModel(p *domain.Product) *ProductModel {
	return &ProductModel{
		ID: p.ID,
		Name: p.Name,
		Description: p.Description,
		Price: p.Price,
		Category: p.Category,
		Rating: p.Rating,
		DeliveryDays: p.DeliveryDays,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

