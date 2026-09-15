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
	links  []PaymentOrderLink
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

func (r *fakeRepository) CreateOrderLinks(ctx context.Context, links []PaymentOrderLink) error {
	r.links = links
	return nil
}

func (r *fakeRepository) FindByIDAndUser(ctx context.Context, paymentID string, userID uint64) (model.Payment, error) {
	return model.Payment{ID: paymentID, UserID: userID, Status: "UNPAID"}, nil
}

func (r *fakeRepository) ListOrders(ctx context.Context, paymentID string, userID uint64) ([]PaymentOrderView, error) {
	return []PaymentOrderView{{OrderID: "order-001", Status: "WAITING_PAYMENT", TicketPrice: "99.00"}}, nil
}

func (r *fakeRepository) StartPay(ctx context.Context, paymentID string, userID uint64) (int64, error) {
	return 1, nil
}

func (r *fakeRepository) MarkSuccess(ctx context.Context, paymentID string, userID uint64, paidAt time.Time) (int64, error) {
	return 1, nil
}

func (r *fakeRepository) MarkFailed(ctx context.Context, paymentID string, userID uint64) (int64, error) {
	return 1, nil
}

func (r *fakeRepository) AddRefundedAmountForOrder(ctx context.Context, orderID string, userID uint64, amount string) (int64, error) {
	return 1, nil
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

func TestCreateOrderLinksAcceptsOneOrTwoLinks(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	err := service.CreateOrderLinks(context.Background(), []PaymentOrderLink{
		{PaymentID: "payment-001", OrderID: "order-001", Amount: "99.00"},
		{PaymentID: "payment-001", OrderID: "order-002", Amount: "99.00"},
	})
	if err != nil {
		t.Fatalf("create order links failed: %v", err)
	}
	if len(repository.links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(repository.links))
	}
}

func TestCreateOrderLinksRejectsInvalidLinks(t *testing.T) {
	err := NewService(&fakeRepository{}).CreateOrderLinks(context.Background(), []PaymentOrderLink{
		{PaymentID: "payment-001", OrderID: "", Amount: "99.00"},
	})
	if !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("expected invalid payment error, got %v", err)
	}
}

func TestStartPayRejectsInvalidRequest(t *testing.T) {
	err := NewService(&fakeRepository{}).StartPay(context.Background(), "", 10001)
	if !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("expected invalid payment error, got %v", err)
	}
}

func TestListOrdersRejectsInvalidRequest(t *testing.T) {
	_, err := NewService(&fakeRepository{}).ListOrders(context.Background(), "", 10001)
	if !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("expected invalid payment error, got %v", err)
	}
}

func TestMarkFailedRejectsInvalidRequest(t *testing.T) {
	err := NewService(&fakeRepository{}).MarkFailed(context.Background(), "", 10001)
	if !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("expected invalid payment error, got %v", err)
	}
}

func TestAddRefundedAmountForOrderRejectsInvalidRequest(t *testing.T) {
	err := NewService(&fakeRepository{}).AddRefundedAmountForOrder(context.Background(), "order-001", 10001, "abc")
	if !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("expected invalid payment error, got %v", err)
	}
}
