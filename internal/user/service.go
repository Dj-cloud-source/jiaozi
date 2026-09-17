package user

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"jiaozi/internal/model"
)

type repository interface {
	FindByID(ctx context.Context, id uint64) (model.User, error)
	AdminList(ctx context.Context, page int, pageSize int) ([]model.User, error)
	AdminFindByID(ctx context.Context, id uint64) (model.User, error)
	UpdatePasswordHash(ctx context.Context, userID uint64, passwordHash string) (int64, error)
}

type Service struct {
	repository repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uint64) (model.User, error) {
	return s.repository.FindByID(ctx, userID)
}

func (s *Service) AdminList(ctx context.Context, req AdminListUsersRequest) ([]model.User, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return s.repository.AdminList(ctx, page, pageSize)
}

func (s *Service) AdminDetail(ctx context.Context, userID uint64) (model.User, error) {
	if userID == 0 {
		return model.User{}, ErrUserNotFound
	}

	return s.repository.AdminFindByID(ctx, userID)
}

func (s *Service) ResetPassword(ctx context.Context, userID uint64, req ResetPasswordRequest) error {
	if userID == 0 {
		return ErrUserNotFound
	}

	newPassword := strings.TrimSpace(req.NewPassword)
	if len(newPassword) < 6 {
		return ErrInvalidUser
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	affected, err := s.repository.UpdatePasswordHash(ctx, userID, string(passwordHash))
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrUserNotFound
	}

	return nil
}
