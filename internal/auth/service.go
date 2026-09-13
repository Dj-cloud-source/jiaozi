package auth

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"jiaozi/internal/model"
)

type userRepository interface {
	CreateUser(ctx context.Context, params CreateUserParams) (model.User, error)
}

type Service struct {
	repository userRepository
}

func NewService(repository userRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (model.User, error) {
	phone := strings.TrimSpace(req.Phone)
	password := strings.TrimSpace(req.Password)
	nickname := strings.TrimSpace(req.Nickname)

	if phone == "" || len(password) < 6 || nickname == "" {
		return model.User{}, ErrInvalidRegisterRequest
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	return s.repository.CreateUser(ctx, CreateUserParams{
		Phone:        phone,
		PasswordHash: string(passwordHash),
		Nickname:     nickname,
	})
}
