package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/domain"
	"github.com/intern/inventory-api/internal/dto"
	"github.com/intern/inventory-api/internal/pkg/password"
	"github.com/intern/inventory-api/internal/pkg/token"
	"github.com/intern/inventory-api/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
}

type authService struct {
	userRepo     repository.UserRepository
	tokenManager *token.Manager
}

func NewAuthService(userRepo repository.UserRepository, tokenManager *token.Manager) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	cleanName := strings.TrimSpace(req.Name)

	existing, err := s.userRepo.FindByEmail(ctx, cleanEmail)
	if err == nil && existing != nil {
		return nil, domain.ErrEmailTaken
	}

	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Name:         cleanName,
		Email:        cleanEmail,
		PasswordHash: hashedPassword,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.FindByEmail(ctx, cleanEmail)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := password.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.tokenManager.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   s.tokenManager.ExpiresIn(),
	}, nil
}

