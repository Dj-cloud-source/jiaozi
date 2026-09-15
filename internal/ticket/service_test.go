package ticket

import (
	"context"
	"errors"
	"testing"
	"time"

	"jiaozi/internal/order"
)

type fakeOrderService struct {
	orderView order.OrderView
	err       error
}

func (s *fakeOrderService) Detail(ctx context.Context, orderID string, userID uint64) (order.OrderView, error) {
	return s.orderView, s.err
}

func TestDetailAllowsTicketStatuses(t *testing.T) {
	for _, status := range []string{"TICKETED", "RETURNED", "COMPLETED"} {
		service := NewService(&fakeOrderService{orderView: order.OrderView{Status: status}})

		_, err := service.Detail(context.Background(), "order-001", 10001)
		if err != nil {
			t.Fatalf("expected status %s to be allowed, got %v", status, err)
		}
	}
}

func TestDetailRejectsUnavailableStatuses(t *testing.T) {
	for _, status := range []string{"WAITING_PAYMENT", "CANCELLED"} {
		service := NewService(&fakeOrderService{orderView: order.OrderView{Status: status}})

		_, err := service.Detail(context.Background(), "order-001", 10001)
		if !errors.Is(err, ErrTicketUnavailable) {
			t.Fatalf("expected unavailable ticket for status %s, got %v", status, err)
		}
	}
}

func TestNewResponseMarksOnlyTicketedValid(t *testing.T) {
	ticketedAt := time.Date(2026, 9, 20, 10, 2, 15, 0, time.Local)
	response := NewResponse(order.OrderView{
		OrderID:              "order-001",
		Status:               "TICKETED",
		TrainNo:              "G101",
		DepartureStationName: "南京南",
		ArrivalStationName:   "上海虹桥",
		DepartureTime:        time.Date(2026, 9, 20, 8, 30, 0, 0, time.Local),
		ArrivalTime:          time.Date(2026, 9, 20, 10, 10, 0, 0, time.Local),
		PassengerName:        "张三",
		PassengerIDCard:      "320101200001011234",
		SeatClass:            "SECOND_CLASS",
		SeatNo:               21,
		TicketPrice:          "99.00",
		TicketedAt:           &ticketedAt,
	})

	if !response.Valid {
		t.Fatal("expected ticketed response to be valid")
	}
	if response.Passenger.IDCardMasked != "3201********1234" {
		t.Fatalf("unexpected masked id card: %s", response.Passenger.IDCardMasked)
	}
	if response.Seat.Class != "SECOND_CLASS" || response.Seat.Number != 21 {
		t.Fatalf("unexpected seat: %+v", response.Seat)
	}
}
