package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/dto"
	"github.com/intern/inventory-api/internal/handler"
	"github.com/intern/inventory-api/internal/middleware"
	"github.com/intern/inventory-api/internal/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockItemService struct {
	mock.Mock
}

func (m *MockItemService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateItemRequest) (*dto.ItemResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ItemResponse), args.Error(1)
}

func (m *MockItemService) Get(ctx context.Context, id uuid.UUID) (*dto.ItemResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ItemResponse), args.Error(1)
}

func (m *MockItemService) List(ctx context.Context, query *dto.ListQuery) ([]dto.ItemResponse, *dto.Meta, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]dto.ItemResponse), args.Get(1).(*dto.Meta), args.Error(2)
}

func (m *MockItemService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateItemRequest) (*dto.ItemResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ItemResponse), args.Error(1)
}

func (m *MockItemService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockItemService) AdjustStock(ctx context.Context, id uuid.UUID, delta int) (*dto.ItemResponse, error) {
	args := m.Called(ctx, id, delta)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ItemResponse), args.Error(1)
}

func setupItemRouter(svc *MockItemService, tokenManager *token.Manager) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewItemHandler(svc)

	itemsGroup := r.Group("/items")
	itemsGroup.Use(middleware.Auth(tokenManager))
	{
		itemsGroup.POST("", h.Create)
		itemsGroup.GET("", h.List)
		itemsGroup.GET("/:id", h.Get)
		itemsGroup.PUT("/:id", h.Update)
		itemsGroup.DELETE("/:id", h.Delete)
		itemsGroup.PATCH("/:id/stock", h.AdjustStock)
	}

	return r
}

func TestItemHandler_AuthAndValidation(t *testing.T) {
	tokenManager := token.NewManager("testsecretwithmorethan32characterslong!!", 24*time.Hour)
	svc := new(MockItemService)
	router := setupItemRouter(svc, tokenManager)

	t.Run("fails without Bearer token - 401", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/items", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("fails with invalid UUID in path - 400", func(t *testing.T) {
		userID := uuid.New()
		jwtToken, err := tokenManager.Generate(userID)
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodGet, "/items/not-a-valid-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails when stock is negative - 422", func(t *testing.T) {
		userID := uuid.New()
		jwtToken, err := tokenManager.Generate(userID)
		assert.NoError(t, err)

		body := []byte(`{"name": "Barang Test", "stock": -5, "price": 10000}`)
		req, _ := http.NewRequest(http.MethodPost, "/items", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

