package service

import (
	"context"
	"fmt"
	"time"

	"github.com/byorty/test-marketplace/services/product-service/internal/domain"
	api "github.com/byorty/test-marketplace/services/product-service/internal/generated"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductService struct {
	repo domain.ProductRepository
	log *zap.Logger
}

func New(log *zap.Logger, repo domain.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
		log: log.Named("product-service"),
	}
}

func (s *ProductService) Create(ctx context.Context, input *api.ProductCreateRequest) (*domain.Product, error) {
	start := time.Now()

	if input == nil {
		s.log.Error(
			"create product failed",
			zap.Error(ErrNilInput),
		)

		return nil, ErrNilInput
	}

	if input.Name == "" {
		s.log.Error(
			"create product failed",
			zap.Error(ErrInvalidProductName),
		)

		return nil, ErrInvalidProductName
	}

	now := time.Now()

	p := &domain.Product{
		ID:           uuid.New(),
		Name:         input.Name,
		Description:  input.Description,
		Category:     input.Category,
		Price:        input.Price,
		DeliveryDays: input.DeliveryDays,
		Rating:       0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, p); err != nil {

		s.log.Error(
			"create product failed",
			zap.Error(err),
			zap.String("product_id", p.ID.String()),
			zap.String("name", p.Name),
		)

		return nil, fmt.Errorf("create product: %w", err)
	}

	s.log.Info(
		"product created",
		zap.String("product_id", p.ID.String()),
		zap.String("name", p.Name),
		zap.Int64("price", p.Price),
		zap.Duration("duration", time.Since(start)),
	)

	return p, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	if id == uuid.Nil {
		s.log.Error(
			"get product failed",
			zap.Error(ErrInvalidID),
		)

		return nil, ErrInvalidID
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(
			"get product failed",
			zap.Error(err),
			zap.String("product_id", id.String()),
		)

		return nil, fmt.Errorf("get product: %w", err)
	}

	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()

	if id == uuid.Nil {
		s.log.Error(
			"delete product failed",
			zap.Error(ErrInvalidID),
		)

		return ErrInvalidID
	}

	if err := s.repo.Delete(ctx, id); err != nil {

		s.log.Error(
			"delete product failed",
			zap.Error(err),
			zap.String("product_id", id.String()),
		)

		return fmt.Errorf("delete product: %w", err)
	}

	s.log.Info(
		"product deleted",
		zap.String("product_id", id.String()),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}

func (s *ProductService) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	res, err := s.repo.List(ctx, filter)
	if err != nil {
		s.log.Error(
			"list products failed",
			zap.Error(err),
			zap.Int("page", filter.Page),
			zap.Int("page_size", filter.PageSize),
		)

		return nil, fmt.Errorf("list products: %w", err)
	}

	return res, nil
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, input *api.ProductUpdateRequest) (*domain.Product, error) {
	start := time.Now()

	if id == uuid.Nil {
		s.log.Error(
			"update product failed",
			zap.Error(ErrInvalidID),
		)

		return nil, ErrInvalidID
	}

	if input == nil {
		s.log.Error(
			"update product failed",
			zap.Error(ErrNilInput),
		)

		return nil, ErrNilInput
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(
			"update product failed",
			zap.Error(err),
			zap.String("product_id", id.String()),
		)

		return nil, fmt.Errorf("get product: %w", err)
	}

	changed := 0

	if input.Name != nil {
		existing.Name = *input.Name
		changed++
	}

	if input.Description != nil {
		existing.Description = *input.Description
		changed++
	}

	if input.Category != nil {
		existing.Category = *input.Category
		changed++
	}

	if input.Price != nil {
		existing.Price = *input.Price
		changed++
	}

	if input.DeliveryDays != nil {
		existing.DeliveryDays = *input.DeliveryDays
		changed++
	}

	if changed == 0 {
		s.log.Error(
			"update product failed",
			zap.Error(ErrEmptyUpdate),
			zap.String("product_id", id.String()),
		)

		return nil, ErrEmptyUpdate
	}

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		s.log.Error(
			"update product failed",
			zap.Error(err),
			zap.String("product_id", id.String()),
		)

		return nil, fmt.Errorf("update product: %w", err)
	}

	s.log.Info(
		"product updated",
		zap.String("product_id", id.String()),
		zap.Duration("duration", time.Since(start)),
	)

	return updated, nil
}