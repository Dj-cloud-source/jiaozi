package passenger

import (
	"context"
	"strings"

	"jiaozi/internal/model"
)

type repository interface {
	Create(ctx context.Context, params CreatePassengerParams) (model.Passenger, error)
	HasActiveTicket(ctx context.Context, idCard string, trainID uint64) (bool, error)
}

type Service struct {
	repository repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, params CreatePassengerParams) (model.Passenger, error) {
	params.Name = strings.TrimSpace(params.Name)
	params.IDCard = strings.TrimSpace(params.IDCard)

	if params.UserID == 0 || params.Name == "" || params.IDCard == "" {
		return model.Passenger{}, ErrInvalidPassenger
	}

	return s.repository.Create(ctx, params)
}

func (s *Service) HasActiveTicket(ctx context.Context, idCard string, trainID uint64) (bool, error) {
	idCard = strings.TrimSpace(idCard)
	if idCard == "" || trainID == 0 {
		return false, ErrInvalidPassenger
	}

	return s.repository.HasActiveTicket(ctx, idCard, trainID)
}
