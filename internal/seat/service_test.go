package seat

import (
	"context"
	"errors"
	"testing"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	seatNos  []uint64
	seats    []model.Seat
	locked   int64
	released int64
	sold     int64
}

func (r *fakeRepository) ListAvailableSeatNos(ctx context.Context, trainID uint64, seatClass string) ([]uint64, error) {
	return r.seatNos, nil
}

func (r *fakeRepository) ListAvailableSeats(ctx context.Context, trainID uint64, seatClass string) ([]model.Seat, error) {
	return r.seats, nil
}

func (r *fakeRepository) LockSeats(ctx context.Context, params LockSeatsParams) (int64, error) {
	return r.locked, nil
}

func (r *fakeRepository) ReleaseSeats(ctx context.Context, params OrderSeatActionParams) (int64, error) {
	return r.released, nil
}

func (r *fakeRepository) MarkSeatsSold(ctx context.Context, params OrderSeatActionParams) (int64, error) {
	return r.sold, nil
}

func TestListAvailableSeatNosReturnsSeatNos(t *testing.T) {
	service := NewService(&fakeRepository{
		seatNos: []uint64{1, 2, 3},
	})

	seatNos, err := service.ListAvailableSeatNos(context.Background(), 10001, "SECOND_CLASS")
	if err != nil {
		t.Fatalf("list available seat nos failed: %v", err)
	}

	assertSeatNos(t, seatNos, []uint64{1, 2, 3})
}

func TestListAvailableSeatsReturnsSeats(t *testing.T) {
	service := NewService(&fakeRepository{
		seats: []model.Seat{
			{ID: 11, TrainID: 10001, SeatClass: "SECOND_CLASS", SeatNo: 1, Status: "AVAILABLE"},
		},
	})

	seats, err := service.ListAvailableSeats(context.Background(), 10001, "SECOND_CLASS")
	if err != nil {
		t.Fatalf("list available seats failed: %v", err)
	}
	if len(seats) != 1 {
		t.Fatalf("expected 1 seat, got %d", len(seats))
	}
	if seats[0].ID != 11 {
		t.Fatalf("expected seat id 11, got %d", seats[0].ID)
	}
}

func TestLockSeatsReturnsAffectedRows(t *testing.T) {
	service := NewService(&fakeRepository{
		locked: 2,
	})

	locked, err := service.LockSeats(context.Background(), LockSeatsParams{
		TrainID:   10001,
		SeatClass: "SECOND_CLASS",
		SeatNos:   []uint64{21, 22},
		OrderID:   "order-001",
	})
	if err != nil {
		t.Fatalf("lock seats failed: %v", err)
	}
	if locked != 2 {
		t.Fatalf("expected 2 locked seats, got %d", locked)
	}
}

func TestLockSeatsRejectsInvalidParams(t *testing.T) {
	_, err := NewService(&fakeRepository{}).LockSeats(context.Background(), LockSeatsParams{
		TrainID:   10001,
		SeatClass: "SECOND_CLASS",
		SeatNos:   []uint64{21},
	})
	if !errors.Is(err, ErrInvalidSeatLock) {
		t.Fatalf("expected invalid seat lock error, got %v", err)
	}
}

func TestReleaseSeatsReturnsAffectedRows(t *testing.T) {
	service := NewService(&fakeRepository{released: 2})

	released, err := service.ReleaseSeats(context.Background(), OrderSeatActionParams{
		OrderID: "order-001",
	})
	if err != nil {
		t.Fatalf("release seats failed: %v", err)
	}
	if released != 2 {
		t.Fatalf("expected 2 released seats, got %d", released)
	}
}

func TestMarkSeatsSoldReturnsAffectedRows(t *testing.T) {
	service := NewService(&fakeRepository{sold: 2})

	sold, err := service.MarkSeatsSold(context.Background(), OrderSeatActionParams{
		OrderID: "order-001",
	})
	if err != nil {
		t.Fatalf("mark seats sold failed: %v", err)
	}
	if sold != 2 {
		t.Fatalf("expected 2 sold seats, got %d", sold)
	}
}

func TestOrderSeatActionsRejectEmptyOrderID(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.ReleaseSeats(context.Background(), OrderSeatActionParams{})
	if !errors.Is(err, ErrInvalidSeatAction) {
		t.Fatalf("expected invalid seat action error, got %v", err)
	}

	_, err = service.MarkSeatsSold(context.Background(), OrderSeatActionParams{})
	if !errors.Is(err, ErrInvalidSeatAction) {
		t.Fatalf("expected invalid seat action error, got %v", err)
	}
}
