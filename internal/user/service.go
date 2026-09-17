package user

import (
	"context"

	"jiaozi/internal/model"
)

type repository interface {
	FindByID(ctx context.Context, id uint64) (model.User, error)
	AdminList(ctx context.Context, page int, pageSize int) ([]model.User, error)
	AdminFindByID(ctx context.Context, id uint64) (model.User, error)
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
