package seat

import (
	"errors"
	"reflect"
	"testing"
)

func TestAllocateOneSeatReturnsSmallestSeatNo(t *testing.T) {
	seats, err := AllocateSeats([]uint64{8, 3, 15}, 1)
	if err != nil {
		t.Fatalf("allocate seats failed: %v", err)
	}

	assertSeatNos(t, seats, []uint64{3})
}

func TestAllocateTwoSeatsPrefersSmallestContinuousPair(t *testing.T) {
	seats, err := AllocateSeats([]uint64{8, 4, 15, 9, 3}, 2)
	if err != nil {
		t.Fatalf("allocate seats failed: %v", err)
	}

	assertSeatNos(t, seats, []uint64{3, 4})
}

func TestAllocateTwoSeatsFallsBackToSmallestTwoSeatNos(t *testing.T) {
	seats, err := AllocateSeats([]uint64{15, 3, 8}, 2)
	if err != nil {
		t.Fatalf("allocate seats failed: %v", err)
	}

	assertSeatNos(t, seats, []uint64{3, 8})
}

func TestAllocateSeatsRejectsInvalidCount(t *testing.T) {
	_, err := AllocateSeats([]uint64{1, 2, 3}, 3)
	if !errors.Is(err, ErrInvalidSeatCount) {
		t.Fatalf("expected invalid seat count error, got %v", err)
	}
}

func TestAllocateSeatsRejectsInsufficientSeats(t *testing.T) {
	_, err := AllocateSeats([]uint64{1}, 2)
	if !errors.Is(err, ErrInsufficientSeats) {
		t.Fatalf("expected insufficient seats error, got %v", err)
	}
}

func TestAllocateSeatsDoesNotMutateInput(t *testing.T) {
	input := []uint64{8, 3, 15}

	_, err := AllocateSeats(input, 1)
	if err != nil {
		t.Fatalf("allocate seats failed: %v", err)
	}

	assertSeatNos(t, input, []uint64{8, 3, 15})
}

func TestAllocateSeatsIgnoresDuplicateSeatNos(t *testing.T) {
	seats, err := AllocateSeats([]uint64{4, 3, 3, 8}, 2)
	if err != nil {
		t.Fatalf("allocate seats failed: %v", err)
	}

	assertSeatNos(t, seats, []uint64{3, 4})
}

func assertSeatNos(t *testing.T, actual []uint64, expected []uint64) {
	t.Helper()

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected seats %v, got %v", expected, actual)
	}
}
