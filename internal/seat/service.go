package seat

import (
	"context"

	"jiaozi/internal/model"
)

type repository interface {
	ListAvailableSeatNos(ctx context.Context, trainID uint64, seatClass string) ([]uint64, error)
	ListAvailableSeats(ctx context.Context, trainID uint64, seatClass string) ([]model.Seat, error)
	LockSeats(ctx context.Context, params LockSeatsParams) (int64, error)
	ReleaseSeats(ctx context.Context, params OrderSeatActionParams) (int64, error)
	MarkSeatsSold(ctx context.Context, params OrderSeatActionParams) (int64, error)
}

type Service struct {
	repository repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListAvailableSeatNos(ctx context.Context, trainID uint64, seatClass string) ([]uint64, error) {
	return s.repository.ListAvailableSeatNos(ctx, trainID, seatClass)
}

func (s *Service) ListAvailableSeats(ctx context.Context, trainID uint64, seatClass string) ([]model.Seat, error) {
	if trainID == 0 || seatClass == "" {
		return nil, ErrInvalidSeatLock
	}

	return s.repository.ListAvailableSeats(ctx, trainID, seatClass)
}

func (s *Service) LockSeats(ctx context.Context, params LockSeatsParams) (int64, error) {
	if params.TrainID == 0 || params.SeatClass == "" || len(params.SeatNos) == 0 || params.OrderID == "" {
		return 0, ErrInvalidSeatLock
	}

	return s.repository.LockSeats(ctx, params)
}

func (s *Service) ReleaseSeats(ctx context.Context, params OrderSeatActionParams) (int64, error) {
	if params.OrderID == "" {
		return 0, ErrInvalidSeatAction
	}

	return s.repository.ReleaseSeats(ctx, params)
}

func (s *Service) MarkSeatsSold(ctx context.Context, params OrderSeatActionParams) (int64, error) {
	if params.OrderID == "" {
		return 0, ErrInvalidSeatAction
	}

	return s.repository.MarkSeatsSold(ctx, params)
}
