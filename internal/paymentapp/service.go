package paymentapp

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
	"jiaozi/internal/order"
	"jiaozi/internal/payment"
	"jiaozi/internal/seat"
)

type Service struct {
	db                *sqlx.DB
	paymentRepository *payment.Repository
	orderRepository   *order.Repository
	seatRepository    *seat.Repository
}

type MockSuccessResult struct {
	Payment model.Payment
	OrderIDs []string
}

func NewService(
	db *sqlx.DB,
	paymentRepository *payment.Repository,
	orderRepository *order.Repository,
	seatRepository *seat.Repository,
) *Service {
	return &Service{
		db:                db,
		paymentRepository: paymentRepository,
		orderRepository:   orderRepository,
		seatRepository:    seatRepository,
	}
}

func (s *Service) Detail(ctx context.Context, paymentID string, userID uint64) (model.Payment, error) {
	return payment.NewService(s.paymentRepository).Detail(ctx, paymentID, userID)
}

func (s *Service) StartPay(ctx context.Context, paymentID string, userID uint64) (model.Payment, error) {
	paymentService := payment.NewService(s.paymentRepository)
	if _, err := paymentService.Detail(ctx, paymentID, userID); err != nil {
		return model.Payment{}, err
	}

	if err := paymentService.StartPay(ctx, paymentID, userID); err != nil {
		return model.Payment{}, err
	}

	return paymentService.Detail(ctx, paymentID, userID)
}

func (s *Service) MockSuccess(ctx context.Context, paymentID string, userID uint64) (MockSuccessResult, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return MockSuccessResult{}, err
	}
	defer tx.Rollback()

	paymentService := payment.NewService(s.paymentRepository.WithExecutor(tx))
	orderService := order.NewService(s.orderRepository.WithExecutor(tx))
	seatService := seat.NewService(s.seatRepository.WithExecutor(tx))

	if _, err := paymentService.Detail(ctx, paymentID, userID); err != nil {
		return MockSuccessResult{}, err
	}

	orderIDs, err := orderService.ListWaitingPaymentIDsByPayment(ctx, paymentID, userID)
	if err != nil {
		return MockSuccessResult{}, err
	}
	if len(orderIDs) == 0 {
		return MockSuccessResult{}, order.ErrOrderStatusInvalid
	}

	paidAt := time.Now()
	if err := paymentService.MarkSuccess(ctx, paymentID, userID, paidAt); err != nil {
		return MockSuccessResult{}, err
	}

	if err := orderService.MarkPaymentOrdersTicketed(ctx, paymentID, userID, paidAt, len(orderIDs)); err != nil {
		return MockSuccessResult{}, err
	}

	for _, orderID := range orderIDs {
		sold, err := seatService.MarkSeatsSold(ctx, seat.OrderSeatActionParams{OrderID: orderID})
		if err != nil {
			return MockSuccessResult{}, err
		}
		if sold != 1 {
			return MockSuccessResult{}, seat.ErrInvalidSeatAction
		}
	}

	if err := tx.Commit(); err != nil {
		return MockSuccessResult{}, err
	}

	paymentModel, err := payment.NewService(s.paymentRepository).Detail(ctx, paymentID, userID)
	if err != nil {
		return MockSuccessResult{}, err
	}

	return MockSuccessResult{
		Payment: paymentModel,
		OrderIDs: orderIDs,
	}, nil
}
