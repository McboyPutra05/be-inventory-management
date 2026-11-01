package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/domain"
)

type CreateItemRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=150"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	Stock       *int   `json:"stock" binding:"required,gte=0"`
	Price       *int64 `json:"price" binding:"required,gte=0"`
}

type UpdateItemRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=150"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	Stock       *int   `json:"stock" binding:"required,gte=0"`
	Price       *int64 `json:"price" binding:"required,gte=0"`
}

type AdjustStockRequest struct {
	Delta int `json:"delta" binding:"required,ne=0"`
}

type ItemResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Stock       int        `json:"stock"`
	Price       int64      `json:"price"`
	CreatedBy   *uuid.UUID `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func ToItemResponse(item *domain.Item) ItemResponse {
	return ItemResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Stock:       item.Stock,
		Price:       item.Price,
		CreatedBy:   item.CreatedBy,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func ToItemResponseList(items []domain.Item) []ItemResponse {
	res := make([]ItemResponse, len(items))
	for i := range items {
		res[i] = ToItemResponse(&items[i])
	}
	return res
}

