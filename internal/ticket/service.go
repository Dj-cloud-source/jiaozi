package ticket

import (
	"context"

	"jiaozi/internal/order"
)

type orderDetailService interface {
	Detail(ctx context.Context, orderID string, userID uint64) (order.OrderView, error)
}

type Service struct {
	orderService orderDetailService
}

func NewService(orderService orderDetailService) *Service {
	return &Service{orderService: orderService}
}

func (s *Service) Detail(ctx context.Context, orderID string, userID uint64) (order.OrderView, error) {
	orderView, err := s.orderService.Detail(ctx, orderID, userID)
	if err != nil {
		return order.OrderView{}, err
	}

	switch orderView.Status {
	case "TICKETED", "RETURNED", "COMPLETED":
		return orderView, nil
	default:
		return order.OrderView{}, ErrTicketUnavailable
	}
}
