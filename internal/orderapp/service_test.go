package orderapp

import (
	"errors"
	"testing"

	"jiaozi/internal/model"
	"jiaozi/internal/train"
)

func TestPriceForSeatClassReturnsFirstClassPrice(t *testing.T) {
	price, err := priceForSeatClass(train.TrainView{
		FirstClassPrice:          "128.00",
		FirstClassAvailableCount: 1,
	}, "FIRST_CLASS")
	if err != nil {
		t.Fatalf("price for first class failed: %v", err)
	}
	if price != "128.00" {
		t.Fatalf("expected 128.00, got %s", price)
	}
}

func TestPriceForSeatClassRejectsSoldOutClass(t *testing.T) {
	_, err := priceForSeatClass(train.TrainView{
		SecondClassPrice:          "88.00",
		SecondClassAvailableCount: 0,
	}, "SECOND_CLASS")
	if !errors.Is(err, ErrNoAvailableSeat) {
		t.Fatalf("expected no available seat error, got %v", err)
	}
}

func TestSelectSeatUsesAllocatorResult(t *testing.T) {
	selectedSeat, err := selectSeat([]model.Seat{
		{ID: 13, SeatNo: 3},
		{ID: 11, SeatNo: 1},
		{ID: 12, SeatNo: 2},
	})
	if err != nil {
		t.Fatalf("select seat failed: %v", err)
	}
	if selectedSeat.ID != 11 {
		t.Fatalf("expected seat id 11, got %d", selectedSeat.ID)
	}
}

func TestSelectSeatRejectsEmptySeats(t *testing.T) {
	_, err := selectSeat(nil)
	if !errors.Is(err, ErrNoAvailableSeat) {
		t.Fatalf("expected no available seat error, got %v", err)
	}
}
