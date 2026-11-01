package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/domain"
	"gorm.io/gorm"
)

type ItemRepository interface {
	Create(ctx context.Context, item *domain.Item) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	List(ctx context.Context, q string, page, limit int) ([]domain.Item, int64, error)
	Update(ctx context.Context, item *domain.Item) error
	Delete(ctx context.Context, id uuid.UUID) error
	AdjustStock(ctx context.Context, id uuid.UUID, delta int) (*domain.Item, error)
}

type itemRepo struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepo{db: db}
}

func (r *itemRepo) Create(ctx context.Context, item *domain.Item) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *itemRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	var item domain.Item
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *itemRepo) List(ctx context.Context, q string, page, limit int) ([]domain.Item, int64, error) {
	var items []domain.Item
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Item{})

	trimmed := strings.TrimSpace(q)
	if trimmed != "" {
		escaped := escapeLike(trimmed)
		query = query.Where("name ILIKE ?", "%"+escaped+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC, id DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *itemRepo) Update(ctx context.Context, item *domain.Item) error {
	item.UpdatedAt = time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&domain.Item{}).
		Where("id = ?", item.ID).
		Select("Name", "Description", "Stock", "Price", "UpdatedAt").
		Updates(item)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *itemRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.Item{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *itemRepo) AdjustStock(ctx context.Context, id uuid.UUID, delta int) (*domain.Item, error) {
	var updated domain.Item

	result := r.db.WithContext(ctx).Raw(`
		UPDATE items
		SET stock = stock + ?, updated_at = now()
		WHERE id = ? AND stock + ? >= 0
		RETURNING id, name, description, stock, price, created_by, created_at, updated_at
	`, delta, id, delta).Scan(&updated)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 || updated.ID == uuid.Nil {
		var count int64
		if err := r.db.WithContext(ctx).Model(&domain.Item{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, domain.ErrNotFound
		}
		return nil, domain.ErrInsufficientStock
	}

	return &updated, nil
}

func escapeLike(query string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(query)
}

