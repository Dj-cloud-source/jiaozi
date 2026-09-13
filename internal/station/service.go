package station

import (
	"context"

	"jiaozi/internal/model"
)

type repository interface {
	ListActive(ctx context.Context) ([]model.Station, error)
}

type Service struct {
	repository repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListActive(ctx context.Context) ([]model.Station, error) {
	return s.repository.ListActive(ctx)
}
