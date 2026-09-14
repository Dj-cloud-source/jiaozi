package order

import (
	"context"
	"time"

	"jiaozi/internal/common"
	"jiaozi/internal/model"
)

type repository interface {
	Create(ctx context.Context, params CreateTicketOrderParams) (model.TicketOrder, error)
	ListByUser(ctx context.Context, userID uint64) ([]OrderView, error)
	FindByIDAndUser(ctx context.Context, orderID string, userID uint64) (OrderView, error)
	CancelWaitingPayment(ctx context.Context, orderID string, userID uint64, cancelledAt time.Time) (int64, error)
}

type Service struct {
	repository repository
}

type CreateTicketOrderRequest struct {
	ID              string
	UserID          uint64
	PassengerID     uint64
	TrainID         uint64
	SeatID          uint64
	TicketPrice     string
	PaymentDeadline time.Time
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, req CreateTicketOrderRequest) (model.TicketOrder, error) {
	if req.UserID == 0 ||
		req.PassengerID == 0 ||
		req.TrainID == 0 ||
		req.SeatID == 0 ||
		req.TicketPrice == "" ||
		req.PaymentDeadline.IsZero() ||
		!common.IsMoney(req.TicketPrice) {
		return model.TicketOrder{}, ErrInvalidTicketOrder
	}

	id := req.ID
	if id == "" {
		generatedID, err := common.NewUUID()
		if err != nil {
			return model.TicketOrder{}, err
		}
		id = generatedID
	}

	return s.repository.Create(ctx, CreateTicketOrderParams{
		ID:              id,
		UserID:          req.UserID,
		PassengerID:     req.PassengerID,
		TrainID:         req.TrainID,
		SeatID:          req.SeatID,
		TicketPrice:     req.TicketPrice,
		Status:          "WAITING_PAYMENT",
		PaymentDeadline: req.PaymentDeadline.Format("2006-01-02 15:04:05"),
	})
}

func parsePaymentDeadline(value string) time.Time {
	deadline, _ := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	return deadline
}

func (s *Service) ListByUser(ctx context.Context, userID uint64) ([]OrderView, error) {
	if userID == 0 {
		return nil, ErrInvalidTicketOrder
	}

	return s.repository.ListByUser(ctx, userID)
}

func (s *Service) Detail(ctx context.Context, orderID string, userID uint64) (OrderView, error) {
	if orderID == "" || userID == 0 {
		return OrderView{}, ErrInvalidTicketOrder
	}

	return s.repository.FindByIDAndUser(ctx, orderID, userID)
}

func (s *Service) CancelWaitingPayment(ctx context.Context, orderID string, userID uint64, cancelledAt time.Time) error {
	if orderID == "" || userID == 0 || cancelledAt.IsZero() {
		return ErrInvalidTicketOrder
	}

	affected, err := s.repository.CancelWaitingPayment(ctx, orderID, userID, cancelledAt)
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrOrderStatusInvalid
	}

	return nil
}
