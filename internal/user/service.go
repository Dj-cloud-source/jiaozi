package user

import (
	"context"

	"jiaozi/internal/model"
)

type repository interface {
	FindByID(ctx context.Context, id uint64) (model.User, error)
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
