package admin

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"jiaozi/internal/auth"
	"jiaozi/internal/model"
)

type repository interface {
	FindByUsername(ctx context.Context, username string) (model.Admin, error)
}

type Service struct {
	repository repository
	tokens     *auth.TokenManager
}

func NewService(repository repository, tokens *auth.TokenManager) *Service {
	return &Service{repository: repository, tokens: tokens}
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
