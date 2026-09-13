package seat

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	seatNos []uint64
	locked  int64
}

func (r *fakeRepository) ListAvailableSeatNos(ctx context.Context, trainID uint64, seatClass string) ([]uint64, error) {
	return r.seatNos, nil
}

func (r *fakeRepository) LockSeats(ctx context.Context, params LockSeatsParams) (int64, error) {
	return r.locked, nil
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
