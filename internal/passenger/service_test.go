package passenger

import (
	"context"
	"errors"
	"testing"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	created         CreatePassengerParams
	hasActiveTicket bool
}

func (r *fakeRepository) Create(ctx context.Context, params CreatePassengerParams) (model.Passenger, error) {
	r.created = params
	return model.Passenger{
		ID:     10001,
		UserID: params.UserID,
		Name:   params.Name,
		IDCard: params.IDCard,
	}, nil
}

func (r *fakeRepository) HasActiveTicket(ctx context.Context, idCard string, trainID uint64) (bool, error) {
	return r.hasActiveTicket, nil
}

func TestCreatePassengerTrimsFields(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	passenger, err := service.Create(context.Background(), CreatePassengerParams{
		UserID: 10001,
		Name:   " 张三 ",
		IDCard: " 320101200001011234 ",
	})
	if err != nil {
		t.Fatalf("create passenger failed: %v", err)
	}

	if passenger.Name != "张三" {
		t.Fatalf("unexpected passenger name: %s", passenger.Name)
	}
	if repository.created.IDCard != "320101200001011234" {
		t.Fatalf("unexpected id card: %s", repository.created.IDCard)
	}
}

func TestCreatePassengerRejectsInvalidParams(t *testing.T) {
	_, err := NewService(&fakeRepository{}).Create(context.Background(), CreatePassengerParams{
		UserID: 10001,
		Name:   " ",
		IDCard: "320101200001011234",
	})
	if !errors.Is(err, ErrInvalidPassenger) {
		t.Fatalf("expected invalid passenger error, got %v", err)
	}
}

func TestHasActiveTicketReturnsRepositoryResult(t *testing.T) {
	service := NewService(&fakeRepository{hasActiveTicket: true})

	hasTicket, err := service.HasActiveTicket(context.Background(), " 320101200001011234 ", 10001)
	if err != nil {
		t.Fatalf("has active ticket failed: %v", err)
	}
	if !hasTicket {
		t.Fatal("expected active ticket")
	}
}

func TestHasActiveTicketRejectsInvalidParams(t *testing.T) {
	_, err := NewService(&fakeRepository{}).HasActiveTicket(context.Background(), "", 10001)
	if !errors.Is(err, ErrInvalidPassenger) {
		t.Fatalf("expected invalid passenger error, got %v", err)
	}
}
