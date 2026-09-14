package payment

import (
	"context"
	"time"

	"jiaozi/internal/common"
	"jiaozi/internal/model"
)

type repository interface {
	Create(ctx context.Context, params CreatePaymentParams) (model.Payment, error)
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
