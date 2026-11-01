package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/domain"
	"github.com/intern/inventory-api/internal/dto"
	"github.com/intern/inventory-api/internal/pkg/password"
	"github.com/intern/inventory-api/internal/pkg/token"
	"github.com/intern/inventory-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	tokenManager := token.NewManager("testsecretwithmorethan32characterslong!!", 24*time.Hour)

	t.Run("success", func(t *testing.T) {
		repo := new(MockUserRepository)
		svc := service.NewAuthService(repo, tokenManager)

		repo.On("FindByEmail", mock.Anything, "user@example.com").Return(nil, domain.ErrNotFound)
		repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

		req := &dto.RegisterRequest{
			Name:     "Test User",
			Email:    "User@Example.com",
			Password: "password123",
		}

		res, err := svc.Register(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "user@example.com", res.Email)
		assert.Equal(t, "Test User", res.Name)
		repo.AssertExpectations(t)
	})

	t.Run("duplicate email", func(t *testing.T) {
		repo := new(MockUserRepository)
		svc := service.NewAuthService(repo, tokenManager)

		existing := &domain.User{
			ID:    uuid.New(),
			Email: "user@example.com",
		}
		repo.On("FindByEmail", mock.Anything, "user@example.com").Return(existing, nil)

		req := &dto.RegisterRequest{
			Name:     "Test User",
			Email:    "user@example.com",
			Password: "password123",
		}

		res, err := svc.Register(context.Background(), req)
		assert.ErrorIs(t, err, domain.ErrEmailTaken)
		assert.Nil(t, res)
		repo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	tokenManager := token.NewManager("testsecretwithmorethan32characterslong!!", 24*time.Hour)

	hashedPassword, err := password.Hash("correct-password")
	assert.NoError(t, err)

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: hashedPassword,
	}

	t.Run("success", func(t *testing.T) {
		repo := new(MockUserRepository)
		svc := service.NewAuthService(repo, tokenManager)

		repo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil)

		req := &dto.LoginRequest{
			Email:    "user@example.com",
			Password: "correct-password",
		}

		res, err := svc.Login(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.Equal(t, "Bearer", res.TokenType)

		extractedID, err := tokenManager.Verify(res.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, extractedID)
	})

	t.Run("wrong password", func(t *testing.T) {
		repo := new(MockUserRepository)
		svc := service.NewAuthService(repo, tokenManager)

		repo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil)

		req := &dto.LoginRequest{
			Email:    "user@example.com",
			Password: "wrong-password",
		}

		res, err := svc.Login(context.Background(), req)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		assert.Nil(t, res)
	})

	t.Run("user not found", func(t *testing.T) {
		repo := new(MockUserRepository)
		svc := service.NewAuthService(repo, tokenManager)

		repo.On("FindByEmail", mock.Anything, "unknown@example.com").Return(nil, domain.ErrNotFound)

		req := &dto.LoginRequest{
			Email:    "unknown@example.com",
			Password: "some-password",
		}

		res, err := svc.Login(context.Background(), req)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		assert.Nil(t, res)
	})
}

