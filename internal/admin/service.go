package admin

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"jiaozi/internal/auth"
	"jiaozi/internal/model"
)

type repository interface {
	Create(ctx context.Context, params CreateAdminParams) (model.Admin, error)
	FindByUsername(ctx context.Context, username string) (model.Admin, error)
}

type Service struct {
	repository repository
	tokens     *auth.TokenManager
}

func NewService(repository repository, tokens *auth.TokenManager) *Service {
	return &Service{repository: repository, tokens: tokens}
}

type CreateAdminRequest struct {
	Username string
	Password string
}

func (s *Service) Create(ctx context.Context, req CreateAdminRequest) (model.Admin, error) {
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || len(password) < 6 {
		return model.Admin{}, ErrInvalidCreateAdminRequest
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.Admin{}, err
	}

	return s.repository.Create(ctx, CreateAdminParams{
		Username:     username,
		PasswordHash: string(passwordHash),
	})
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || password == "" {
		return LoginResponse{}, ErrInvalidLoginRequest
	}

	admin, err := s.repository.FindByUsername(ctx, username)
	if err != nil {
		return LoginResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return LoginResponse{}, ErrInvalidUsernameOrPassword
	}

	token, err := s.tokens.IssueAdmin(admin.ID)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		AccessToken: token,
		ExpiresIn:   s.tokens.ExpiresIn(),
		Admin: LoginAdminData{
			ID:       admin.ID,
			Username: admin.Username,
		},
	}, nil
}
