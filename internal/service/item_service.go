package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/domain"
	"github.com/intern/inventory-api/internal/dto"
	"github.com/intern/inventory-api/internal/repository"
)

type ItemService interface {
	Create(ctx context.Context, userID uuid.UUID, req *dto.CreateItemRequest) (*dto.ItemResponse, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.ItemResponse, error)
	List(ctx context.Context, query *dto.ListQuery) ([]dto.ItemResponse, *dto.Meta, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateItemRequest) (*dto.ItemResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AdjustStock(ctx context.Context, id uuid.UUID, delta int) (*dto.ItemResponse, error)
}

type itemService struct {
	itemRepo repository.ItemRepository
}

func NewItemService(itemRepo repository.ItemRepository) ItemService {
	return &itemService{itemRepo: itemRepo}
}

func (s *itemService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateItemRequest) (*dto.ItemResponse, error) {
	now := time.Now().UTC()
	item := &domain.Item{
		ID:          uuid.New(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Stock:       *req.Stock,
		Price:       *req.Price,
		CreatedBy:   &userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	res := dto.ToItemResponse(item)
	return &res, nil
}

func (s *itemService) Get(ctx context.Context, id uuid.UUID) (*dto.ItemResponse, error) {
	item, err := s.itemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := dto.ToItemResponse(item)
	return &res, nil
}

func (s *itemService) List(ctx context.Context, query *dto.ListQuery) ([]dto.ItemResponse, *dto.Meta, error) {
	query.ApplyDefaults()

	items, total, err := s.itemRepo.List(ctx, query.Q, query.Page, query.Limit)
	if err != nil {
		return nil, nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(query.Limit) - 1) / int64(query.Limit))
	}

	meta := &dto.Meta{
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return dto.ToItemResponseList(items), meta, nil
}

func (s *itemService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateItemRequest) (*dto.ItemResponse, error) {
	existing, err := s.itemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Name = strings.TrimSpace(req.Name)
	existing.Description = strings.TrimSpace(req.Description)
	existing.Stock = *req.Stock
	existing.Price = *req.Price

	if err := s.itemRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	res := dto.ToItemResponse(existing)
	return &res, nil
}

func (s *itemService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.itemRepo.Delete(ctx, id)
}

func (s *itemService) AdjustStock(ctx context.Context, id uuid.UUID, delta int) (*dto.ItemResponse, error) {
	updated, err := s.itemRepo.AdjustStock(ctx, id, delta)
	if err != nil {
		return nil, err
	}
	res := dto.ToItemResponse(updated)
	return &res, nil
}

