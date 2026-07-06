package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	domain "github.com/byorty/test-marketplace/services/product-service/internal/domain/product"
	"github.com/google/uuid"
)

type service struct {
	repo domain.Repository
	log *slog.Logger
}

func New(log *slog.Logger, repo domain.Repository) *service {
	return &service{
		repo: repo,
		log: log.With("layer", "service"),
	}
}

func logError(log *slog.Logger, op string, err error) {
	log.Error("operatoin failed", "op", op, "error", err)
}

func (s *service) Create(ctx context.Context, input *CreateProduct) (*domain.Product, error) {
	const op = "Service.Create"

	start := time.Now()

	if input == nil {
		err := fmt.Errorf("%s: nil input", op)
		logError(s.log, op, err)
		return nil, err
	}

	s.log.Info("create product started", "op", op, "name", input.Name, "price", input.Price)

	if input.Name == "" {
		err := ErrInvalidProductName
		logError(s.log, op, err)
		return nil, err
	}

	now := time.Now()

	p := &domain.Product{
		ID: uuid.New(),
		Name: input.Name,
		Description: input.Description,
		Category: input.Category,
		Price: input.Price,
		DeliveryDays: input.DeliveryDays,
		Rating: 0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		logError(s.log, op, err)
		return nil, fmt.Errorf("%s: create: %w", op, err)
	}
	s.log.Info("create product success", "op", op, "id", p.ID, "duration_ms", time.Since(start).Milliseconds())

	return p, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	const op = "Service.GetByID"

	s.log.Debug("get product", "op", op, "id", id)

	if id == uuid.Nil {
		err := fmt.Errorf("%s: invalid id", op)
		logError(s.log, op, err)
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logError(s.log, op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return p, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "Service.Delete"

	start := time.Now()

	s.log.Info("delete product", "op", op, "id", id)

	if id == uuid.Nil {
		err := fmt.Errorf("%s: invalid id", op)
		logError(s.log, op, err)
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logError(s.log, op, err)
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("delete product success", "op", op, "id", id, "duration_ms", time.Since(start).Milliseconds())

	return nil
}

func (s *service) List(ctx context.Context, filter domain.ListFilter) (*domain.ProductList, error) {
	const op = "Service.List"

	s.log.Debug("list products", "op", op, "filter", filter)

	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	res, err := s.repo.List(ctx, filter)
	if err != nil {
		logError(s.log, op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	s.log.Debug("list products success", "op", op, "total", res.Total, "page", res.Page, "page_size", res.PageSize)

	return res, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, input *UpdateProduct) error {
	const op = "Service.Update"

	start := time.Now()

	if id == uuid.Nil {
		err := fmt.Errorf("%s: invalid id", op)
		logError(s.log, op, err)
		return err
	}

	if input == nil {
		err := fmt.Errorf("%s: input is empty", op)
		logError(s.log, op, err)
		return err
	}

	s.log.Info("update product started", "op", op, "id", id)

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logError(s.log, op, err)
		return fmt.Errorf("%s: get: %w", op, err)
	}

	changed := 0

	if input.Name != nil {
		existing.Name = *input.Name
		changed ++
	}

	if input.Description != nil {
		existing.Description = *input.Description
		changed ++
	}

	if input.Category != nil {
		existing.Category = *input.Category
		changed ++
	}

	if input.Price != nil {
		existing.Price = *input.Price
		changed ++
	}

	if input.DelilveryDays != nil {
		existing.DeliveryDays = *input.DelilveryDays
		changed ++
	}

	if changed == 0 {
		err := ErrEmptyUpdate
		logError(s.log, op, err)
		return err
	}

	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, existing); err != nil {
		logError(s.log, op, err)
		return fmt.Errorf("%s: update: %w", op, err)
	}

	s.log.Info("update product success", "op", op, "id", id, "duration_ms", time.Since(start).Milliseconds())

	return nil
}

