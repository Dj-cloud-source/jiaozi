package payment

import (
	"context"
	"time"

	"jiaozi/internal/common"
	"jiaozi/internal/model"
)

type repository interface {
	Create(ctx context.Context, params CreatePaymentParams) (model.Payment, error)
	CreateOrderLinks(ctx context.Context, links []PaymentOrderLink) error
	FindByIDAndUser(ctx context.Context, paymentID string, userID uint64) (model.Payment, error)
	ListOrders(ctx context.Context, paymentID string, userID uint64) ([]PaymentOrderView, error)
	StartPay(ctx context.Context, paymentID string, userID uint64) (int64, error)
	MarkSuccess(ctx context.Context, paymentID string, userID uint64, paidAt time.Time) (int64, error)
	MarkFailed(ctx context.Context, paymentID string, userID uint64) (int64, error)
	AddRefundedAmountForOrder(ctx context.Context, orderID string, userID uint64, amount string) (int64, error)
}

type Service struct {
	repository repository
}

type CreatePaymentRequest struct {
	UserID          uint64
	Amount          string
	PaymentDeadline time.Time
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, req CreatePaymentRequest) (model.Payment, error) {
	if req.UserID == 0 ||
		req.Amount == "" ||
		req.PaymentDeadline.IsZero() ||
		!common.IsMoney(req.Amount) {
		return model.Payment{}, ErrInvalidPayment
	}

	id, err := common.NewUUID()
	if err != nil {
		return model.Payment{}, err
	}

	return s.repository.Create(ctx, CreatePaymentParams{
		ID:              id,
		UserID:          req.UserID,
		OriginalAmount:  req.Amount,
		PayableAmount:   req.Amount,
		RefundedAmount:  "0.00",
		Status:          "UNPAID",
		PaymentDeadline: req.PaymentDeadline.Format("2006-01-02 15:04:05"),
		Provider:        "MOCK",
	})
}

func (s *Service) CreateOrderLinks(ctx context.Context, links []PaymentOrderLink) error {
	if len(links) == 0 || len(links) > 2 {
		return ErrInvalidPayment
	}

	for _, link := range links {
		if link.PaymentID == "" ||
			link.OrderID == "" ||
			link.Amount == "" ||
			!common.IsMoney(link.Amount) {
			return ErrInvalidPayment
		}
	}

	return s.repository.CreateOrderLinks(ctx, links)
}

func (s *Service) Detail(ctx context.Context, paymentID string, userID uint64) (model.Payment, error) {
	if paymentID == "" || userID == 0 {
		return model.Payment{}, ErrInvalidPayment
	}

	return s.repository.FindByIDAndUser(ctx, paymentID, userID)
}

func (s *Service) ListOrders(ctx context.Context, paymentID string, userID uint64) ([]PaymentOrderView, error) {
	if paymentID == "" || userID == 0 {
		return nil, ErrInvalidPayment
	}

	return s.repository.ListOrders(ctx, paymentID, userID)
}

func (s *Service) StartPay(ctx context.Context, paymentID string, userID uint64) error {
	if paymentID == "" || userID == 0 {
		return ErrInvalidPayment
	}

	affected, err := s.repository.StartPay(ctx, paymentID, userID)
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrPaymentStatusInvalid
	}

	return nil
}

func (s *Service) MarkSuccess(ctx context.Context, paymentID string, userID uint64, paidAt time.Time) error {
	if paymentID == "" || userID == 0 || paidAt.IsZero() {
		return ErrInvalidPayment
	}

	affected, err := s.repository.MarkSuccess(ctx, paymentID, userID, paidAt)
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrPaymentStatusInvalid
	}

	return nil
}

func (s *Service) MarkFailed(ctx context.Context, paymentID string, userID uint64) error {
	if paymentID == "" || userID == 0 {
		return ErrInvalidPayment
	}

	affected, err := s.repository.MarkFailed(ctx, paymentID, userID)
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrPaymentStatusInvalid
	}

	return nil
}

func (s *Service) AddRefundedAmountForOrder(ctx context.Context, orderID string, userID uint64, amount string) error {
	if orderID == "" || userID == 0 || amount == "" || !common.IsMoney(amount) {
		return ErrInvalidPayment
	}

	affected, err := s.repository.AddRefundedAmountForOrder(ctx, orderID, userID, amount)
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrPaymentStatusInvalid
	}

	return nil
}
