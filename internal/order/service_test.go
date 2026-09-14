package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	params CreateTicketOrderParams
	orders []OrderView
	detail OrderView
}

func (r *fakeRepository) Create(ctx context.Context, params CreateTicketOrderParams) (model.TicketOrder, error) {
	r.params = params
	return model.TicketOrder{
		ID:              params.ID,
		UserID:          params.UserID,
		PassengerID:     params.PassengerID,
		TrainID:         params.TrainID,
		SeatID:          params.SeatID,
		TicketPrice:     params.TicketPrice,
		Status:          params.Status,
		PaymentDeadline: parsePaymentDeadline(params.PaymentDeadline),
	}, nil
}

func (r *fakeRepository) ListByUser(ctx context.Context, userID uint64) ([]OrderView, error) {
	return r.orders, nil
}

func (r *fakeRepository) FindByIDAndUser(ctx context.Context, orderID string, userID uint64) (OrderView, error) {
	return r.detail, nil
}

func (r *fakeRepository) CancelWaitingPayment(ctx context.Context, orderID string, userID uint64, cancelledAt time.Time) (int64, error) {
	return 1, nil
}

func (r *fakeRepository) ListWaitingPaymentIDsByPayment(ctx context.Context, paymentID string, userID uint64) ([]string, error) {
	return []string{"order-001"}, nil
}

func (r *fakeRepository) MarkPaymentOrdersTicketed(ctx context.Context, paymentID string, userID uint64, ticketedAt time.Time) (int64, error) {
	return 1, nil
}

func (r *fakeRepository) ListExpiredWaitingPaymentIDs(ctx context.Context, query ExpiredOrderQuery) ([]string, error) {
	return []string{"order-001"}, nil
}

func (r *fakeRepository) CancelExpiredWaitingPayment(ctx context.Context, orderID string, cancelledAt time.Time) (int64, error) {
	return 1, nil
}

func TestCreateTicketOrderCreatesWaitingPaymentOrder(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	deadline := time.Date(2026, 9, 20, 10, 10, 0, 0, time.Local)

	order, err := service.Create(context.Background(), CreateTicketOrderRequest{
		UserID:          10001,
		PassengerID:     20001,
		TrainID:         30001,
		SeatID:          40001,
		TicketPrice:     "99.00",
		PaymentDeadline: deadline,
	})
	if err != nil {
		t.Fatalf("create ticket order failed: %v", err)
	}

	if order.ID == "" {
		t.Fatal("order id is empty")
	}
	if order.Status != "WAITING_PAYMENT" {
		t.Fatalf("unexpected order status: %s", order.Status)
	}
	if repository.params.PaymentDeadline != "2026-09-20 10:10:00" {
		t.Fatalf("unexpected payment deadline: %s", repository.params.PaymentDeadline)
	}
}

func TestCreateTicketOrderRejectsInvalidRequest(t *testing.T) {
	_, err := NewService(&fakeRepository{}).Create(context.Background(), CreateTicketOrderRequest{
		UserID:      10001,
		PassengerID: 20001,
		TrainID:     30001,
		SeatID:      40001,
		TicketPrice: "abc",
	})
	if !errors.Is(err, ErrInvalidTicketOrder) {
		t.Fatalf("expected invalid ticket order error, got %v", err)
	}
}

func TestListByUserReturnsOrders(t *testing.T) {
	service := NewService(&fakeRepository{
		orders: []OrderView{
			{OrderID: "order-001", UserID: 10001},
		},
	})

	orders, err := service.ListByUser(context.Background(), 10001)
	if err != nil {
		t.Fatalf("list orders failed: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
}

func TestDetailRejectsInvalidRequest(t *testing.T) {
	_, err := NewService(&fakeRepository{}).Detail(context.Background(), "", 10001)
	if !errors.Is(err, ErrInvalidTicketOrder) {
		t.Fatalf("expected invalid ticket order error, got %v", err)
	}
}

func TestCancelWaitingPaymentRejectsInvalidRequest(t *testing.T) {
	err := NewService(&fakeRepository{}).CancelWaitingPayment(context.Background(), "", 10001, time.Now())
	if !errors.Is(err, ErrInvalidTicketOrder) {
		t.Fatalf("expected invalid ticket order error, got %v", err)
	}
}

func TestMarkPaymentOrdersTicketedRejectsInvalidRequest(t *testing.T) {
	err := NewService(&fakeRepository{}).MarkPaymentOrdersTicketed(context.Background(), "", 10001, time.Now(), 1)
	if !errors.Is(err, ErrInvalidTicketOrder) {
		t.Fatalf("expected invalid ticket order error, got %v", err)
	}
}

func TestListExpiredWaitingPaymentIDsRejectsInvalidRequest(t *testing.T) {
	_, err := NewService(&fakeRepository{}).ListExpiredWaitingPaymentIDs(context.Background(), time.Now(), 0)
	if !errors.Is(err, ErrInvalidTicketOrder) {
		t.Fatalf("expected invalid ticket order error, got %v", err)
	}
}
