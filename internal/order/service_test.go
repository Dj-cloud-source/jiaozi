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
