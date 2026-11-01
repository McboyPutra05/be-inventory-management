package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/domain"
	"github.com/intern/inventory-api/internal/dto"
	"github.com/intern/inventory-api/internal/handler"
	"github.com/intern/inventory-api/internal/pkg/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RegisterResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func setupAuthRouter(svc *MockAuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewAuthHandler(svc)
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	return r
}

func TestAuthHandler_Register_Validation(t *testing.T) {
	svc := new(MockAuthService)
	router := setupAuthRouter(svc)

	t.Run("fails when required fields are missing - 422", func(t *testing.T) {
		body := []byte(`{"name": ""}`)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp httpx.ResponseError
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, httpx.CodeValidationError, resp.Error.Code)
		assert.NotEmpty(t, resp.Error.Details)
	})

	t.Run("fails with invalid JSON body - 400", func(t *testing.T) {
		body := []byte(`{"name": invalid-json}`)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success returns 201", func(t *testing.T) {
		registered := &dto.RegisterResponse{
			ID:        uuid.New(),
			Name:      "Budi Santoso",
			Email:     "budi@example.com",
			CreatedAt: time.Now(),
		}
		svc.On("Register", mock.Anything, mock.MatchedBy(func(r *dto.RegisterRequest) bool {
			return r.Email == "budi@example.com"
		})).Return(registered, nil)

		body := []byte(`{"name": "Budi Santoso", "email": "budi@example.com", "password": "password123"}`)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp httpx.ResponseSuccess
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Data)
	})

	t.Run("duplicate email returns 409", func(t *testing.T) {
		svc.On("Register", mock.Anything, mock.MatchedBy(func(r *dto.RegisterRequest) bool {
			return r.Email == "existing@example.com"
		})).Return(nil, domain.ErrEmailTaken)

		body := []byte(`{"name": "Budi", "email": "existing@example.com", "password": "password123"}`)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		var resp httpx.ResponseError
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, httpx.CodeEmailAlreadyExists, resp.Error.Code)
	})
}

