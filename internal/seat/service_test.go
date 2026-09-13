package seat

import (
	"context"
	"testing"
)

type fakeRepository struct {
	seatNos []uint64
}

func (r *fakeRepository) ListAvailableSeatNos(ctx context.Context, trainID uint64, seatClass string) ([]uint64, error) {
	return r.seatNos, nil
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
