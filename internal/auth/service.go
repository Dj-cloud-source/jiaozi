package auth

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"jiaozi/internal/model"
)

type userRepository interface {
	CreateUser(ctx context.Context, params CreateUserParams) (model.User, error)
	FindUserByPhone(ctx context.Context, phone string) (model.User, error)
}

type Service struct {
	repository userRepository
	tokens     *TokenManager
}

func NewService(repository userRepository, tokens *TokenManager) *Service {
	return &Service{repository: repository, tokens: tokens}
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

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	phone := strings.TrimSpace(req.Phone)
	password := strings.TrimSpace(req.Password)
	if phone == "" || password == "" {
		return LoginResponse{}, ErrInvalidLoginRequest
	}

	user, err := s.repository.FindUserByPhone(ctx, phone)
	if err != nil {
		return LoginResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginResponse{}, ErrInvalidPhoneOrPassword
	}

	token, err := s.tokens.Issue(user.ID)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		AccessToken: token,
		ExpiresIn:   s.tokens.ExpiresIn(),
		User: LoginUserData{
			ID:       user.ID,
			Phone:    user.Phone,
			Nickname: user.Nickname,
		},
	}, nil
}
