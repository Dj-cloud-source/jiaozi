package station

import (
	"context"
	"strings"

	"jiaozi/internal/model"
)

type repository interface {
	ListActive(ctx context.Context) ([]model.Station, error)
	ListAll(ctx context.Context) ([]model.Station, error)
	FindByID(ctx context.Context, stationID uint64) (model.Station, error)
	Create(ctx context.Context, name string) (model.Station, error)
	UpdateName(ctx context.Context, stationID uint64, name string) (model.Station, error)
	Disable(ctx context.Context, stationID uint64) (model.Station, error)
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

func (s *Service) ListAll(ctx context.Context) ([]model.Station, error) {
	return s.repository.ListAll(ctx)
}

func (s *Service) Detail(ctx context.Context, stationID uint64) (model.Station, error) {
	if stationID == 0 {
		return model.Station{}, ErrStationNotFound
	}

	return s.repository.FindByID(ctx, stationID)
}

func (s *Service) Create(ctx context.Context, req CreateStationRequest) (model.Station, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return model.Station{}, ErrInvalidStationName
	}

	return s.repository.Create(ctx, name)
}

func (s *Service) UpdateName(ctx context.Context, stationID uint64, req UpdateStationRequest) (model.Station, error) {
	if stationID == 0 {
		return model.Station{}, ErrStationNotFound
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return model.Station{}, ErrInvalidStationName
	}

	return s.repository.UpdateName(ctx, stationID, name)
}

func (s *Service) Disable(ctx context.Context, stationID uint64) (model.Station, error) {
	if stationID == 0 {
		return model.Station{}, ErrStationNotFound
	}

	return s.repository.Disable(ctx, stationID)
}
