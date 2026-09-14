package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	params CreatePaymentParams
}

func (r *fakeRepository) Create(ctx context.Context, params CreatePaymentParams) (model.Payment, error) {
	r.params = params
	return model.Payment{
		ID:              params.ID,
		UserID:          params.UserID,
		OriginalAmount:  params.OriginalAmount,
		PayableAmount:   params.PayableAmount,
		RefundedAmount:  params.RefundedAmount,
		Status:          params.Status,
		PaymentDeadline: parsePaymentDeadline(params.PaymentDeadline),
		Provider:        params.Provider,
	}, nil
}

func TestCreatePaymentCreatesUnpaidMockPayment(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	deadline := time.Date(2026, 9, 20, 10, 10, 0, 0, time.Local)

	payment, err := service.Create(context.Background(), CreatePaymentRequest{
		UserID:          10001,
		Amount:          "198.00",
		PaymentDeadline: deadline,
	})
	if err != nil {
		t.Fatalf("create payment failed: %v", err)
	}

	if payment.ID == "" {
		t.Fatal("payment id is empty")
	}
	if payment.Status != "UNPAID" {
		t.Fatalf("unexpected payment status: %s", payment.Status)
	}
	if payment.Provider != "MOCK" {
		t.Fatalf("unexpected provider: %s", payment.Provider)
	}
	if repository.params.RefundedAmount != "0.00" {
		t.Fatalf("unexpected refunded amount: %s", repository.params.RefundedAmount)
	}
}

func TestCreatePaymentRejectsInvalidRequest(t *testing.T) {
	_, err := NewService(&fakeRepository{}).Create(context.Background(), CreatePaymentRequest{
		UserID: 10001,
		Amount: "abc",
	})
	if !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("expected invalid payment error, got %v", err)
	}
}
