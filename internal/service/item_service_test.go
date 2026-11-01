package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/domain"
	"github.com/intern/inventory-api/internal/dto"
	"github.com/intern/inventory-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) Create(ctx context.Context, item *domain.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Item), args.Error(1)
}

func (m *MockItemRepository) List(ctx context.Context, q string, page, limit int) ([]domain.Item, int64, error) {
	args := m.Called(ctx, q, page, limit)
	return args.Get(0).([]domain.Item), args.Get(1).(int64), args.Error(2)
}

func (m *MockItemRepository) Update(ctx context.Context, item *domain.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockItemRepository) AdjustStock(ctx context.Context, id uuid.UUID, delta int) (*domain.Item, error) {
	args := m.Called(ctx, id, delta)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Item), args.Error(1)
}

func TestItemService_CreateAndGet(t *testing.T) {
	repo := new(MockItemRepository)
	svc := service.NewItemService(repo)

	userID := uuid.New()
	stock := 15
	price := int64(50000)

	t.Run("create success", func(t *testing.T) {
		repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Item")).Return(nil)

		req := &dto.CreateItemRequest{
			Name:        "Laptop Stand",
			Description: "Aluminium stand",
			Stock:       &stock,
			Price:       &price,
		}

		res, err := svc.Create(context.Background(), userID, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Laptop Stand", res.Name)
		assert.Equal(t, 15, res.Stock)
		assert.Equal(t, int64(50000), res.Price)
	})

	t.Run("get success", func(t *testing.T) {
		itemID := uuid.New()
		item := &domain.Item{
			ID:          itemID,
			Name:        "Mouse Pad",
			Description: "Smooth desk pad",
			Stock:       10,
			Price:       25000,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		repo.On("FindByID", mock.Anything, itemID).Return(item, nil)

		res, err := svc.Get(context.Background(), itemID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Mouse Pad", res.Name)
	})

	t.Run("get not found", func(t *testing.T) {
		unknownID := uuid.New()
		repo.On("FindByID", mock.Anything, unknownID).Return(nil, domain.ErrNotFound)

		res, err := svc.Get(context.Background(), unknownID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, res)
	})
}

func TestItemService_AdjustStock(t *testing.T) {
	repo := new(MockItemRepository)
	svc := service.NewItemService(repo)
	itemID := uuid.New()

	t.Run("successful stock deduction", func(t *testing.T) {
		updated := &domain.Item{
			ID:    itemID,
			Name:  "Mechanical Keyboard",
			Stock: 9,
			Price: 450000,
		}

		repo.On("AdjustStock", mock.Anything, itemID, -1).Return(updated, nil)

		res, err := svc.AdjustStock(context.Background(), itemID, -1)
		assert.NoError(t, err)
		assert.Equal(t, 9, res.Stock)
	})

	t.Run("insufficient stock error", func(t *testing.T) {
		repo.On("AdjustStock", mock.Anything, itemID, -20).Return(nil, domain.ErrInsufficientStock)

		res, err := svc.AdjustStock(context.Background(), itemID, -20)
		assert.ErrorIs(t, err, domain.ErrInsufficientStock)
		assert.Nil(t, res)
	})
}

