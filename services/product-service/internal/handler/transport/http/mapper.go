package http

import (
	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	s "github.com/byorty/test-marketplace/services/product-service/internal/service"
)

func toCreateInput(req api.ProductCreateRequest) *s.CreateProduct {
	return &s.CreateProduct{
		Category: req.Category,
		Description: req.Description,
		DeliveryDays: req.DeliveryDays,
		Name: req.Name,
		Price: req.Price,
	}
}

func toUpdateInput(req api.ProductUpdateRequest) *s.UpdateProduct {
	update := &s.UpdateProduct{}

	if req.Name != nil {
		update.Name = req.Name
	}

	if req.Description != nil {
		update.Description = req.Description
	}

	if req.DeliveryDays != nil {
		update.DelilveryDays = req.DeliveryDays
	}

	if req.Category != nil {
		update.Category = req.Category
	}

	if req.Price != nil {
		update.Price = req.Price
	}
	return update
}

func toResponse(product *domain.Product) api.ProductResponse {
	return api.ProductResponse{
		Id: product.ID,
		Name: product.Name,
		Description: product.Description,
		DeliveryDays: product.DeliveryDays,
		Category: product.Category,
		Price: product.Price,
		Rating: product.Rating,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}

func toProductList(list *domain.ProductList) api.ProductListResponse {
	items := make([]api.ProductResponse, len(list.Items))

	for i, p := range list.Items {
		items[i] = toResponse(&p)
	}

	return api.ProductListResponse{
		Items: items,
		Total: list.Total,
		Page: list.Page,
		PageSize: list.PageSize,
	}
}

func toListFilter(params api.GetProductsParams) domain.ListFilter {
	filter := domain.ListFilter{
		Page: 1,
		PageSize: 20,
		Order: domain.Asc,
	}

	if params.Name != nil {
		filter.Name = *params.Name
	}

	if params.Category != nil {
		filter.Category = *params.Category
	}

	if params.MinPrice != nil {
		filter.MinPrice = params.MinPrice
	}

	if params.MaxPrice != nil {
		filter.MaxPrice = params.MaxPrice
	}

	if params.MinRating != nil {
		filter.MinRating = params.MinRating
	}

	if params.MaxDeliveryDays != nil {
		filter.MaxDeliveryDays = params.MaxDeliveryDays
	}

	if params.Page != nil {
		filter.Page = *params.Page
	}

	if params.PageSize != nil {
		filter.PageSize = *params.PageSize
	}

	if params.SortBy != nil {
		switch *params.SortBy {
		case api.Price:
			filter.SortBy = domain.SortByPrice

		case api.Rating:
			filter.SortBy = domain.SortByRating
		}
	}

	if params.Order != nil {
		switch *params.Order {
		case api.Desc:
			filter.Order = domain.Desc

		default:
			filter.Order = domain.Asc
		}
	}

	return filter
}