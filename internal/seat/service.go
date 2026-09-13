package seat

import "context"

type repository interface {
	ListAvailableSeatNos(ctx context.Context, trainID uint64, seatClass string) ([]uint64, error)
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
