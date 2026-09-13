package station

import (
	"context"
	"strings"

	"jiaozi/internal/model"
)

type repository interface {
	ListActive(ctx context.Context) ([]model.Station, error)
	Create(ctx context.Context, name string) (model.Station, error)
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

func (s *Service) Create(ctx context.Context, req CreateStationRequest) (model.Station, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return model.Station{}, ErrInvalidStationName
	}

	return s.repository.Create(ctx, name)
}
