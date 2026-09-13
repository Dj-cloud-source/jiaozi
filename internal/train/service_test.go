package train

import (
	"context"
	"errors"
	"testing"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	params CreateTrainParams
	err    error
}

func (r *fakeRepository) Create(ctx context.Context, params CreateTrainParams) (model.Train, error) {
	r.params = params
	if r.err != nil {
		return model.Train{}, r.err
	}

	return model.Train{
		ID:                   10001,
		TrainNo:              params.TrainNo,
		DepartureDate:        params.DepartureDate,
		DepartureStationID:   params.DepartureStationID,
		ArrivalStationID:     params.ArrivalStationID,
		DepartureTime:        params.DepartureTime,
		ArrivalTime:          params.ArrivalTime,
		SaleStartTime:        params.SaleStartTime,
		FirstClassPrice:      params.FirstClassPrice,
		SecondClassPrice:     params.SecondClassPrice,
		FirstClassSeatCount:  params.FirstClassSeatCount,
		SecondClassSeatCount: params.SecondClassSeatCount,
		Status:               "DRAFT",
	}, nil
}

func TestCreateTrainCreatesDraft(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	train, err := service.Create(context.Background(), validCreateTrainRequest())
	if err != nil {
		t.Fatalf("create train failed: %v", err)
	}

	if train.Status != "DRAFT" {
		t.Fatalf("unexpected status: %s", train.Status)
	}
	if repository.params.TrainNo != "G101" {
		t.Fatalf("unexpected train no: %s", repository.params.TrainNo)
	}
	if repository.params.SecondClassPrice != "99.00" {
		t.Fatalf("unexpected second class price: %s", repository.params.SecondClassPrice)
	}
}

func TestCreateTrainRejectsInvalidTrainNo(t *testing.T) {
	req := validCreateTrainRequest()
	req.TrainNo = "g101"

	_, err := NewService(&fakeRepository{}).Create(context.Background(), req)
	if !errors.Is(err, ErrInvalidTrain) {
		t.Fatalf("expected invalid train error, got %v", err)
	}
}

func TestCreateTrainRejectsSameStations(t *testing.T) {
	req := validCreateTrainRequest()
	req.ArrivalStationID = req.DepartureStationID

	_, err := NewService(&fakeRepository{}).Create(context.Background(), req)
	if !errors.Is(err, ErrInvalidTrain) {
		t.Fatalf("expected invalid train error, got %v", err)
	}
}

func TestCreateTrainRejectsCrossDayTrain(t *testing.T) {
	req := validCreateTrainRequest()
	req.ArrivalTime = "2026-09-21T10:10:00+08:00"

	_, err := NewService(&fakeRepository{}).Create(context.Background(), req)
	if !errors.Is(err, ErrInvalidTrain) {
		t.Fatalf("expected invalid train error, got %v", err)
	}
}

func TestCreateTrainClearsPriceWhenSeatCountIsZero(t *testing.T) {
	repository := &fakeRepository{}
	req := validCreateTrainRequest()
	req.FirstClassSeatCount = 0
	req.FirstClassPrice = "199.00"

	_, err := NewService(repository).Create(context.Background(), req)
	if err != nil {
		t.Fatalf("create train failed: %v", err)
	}

	if repository.params.FirstClassPrice != "" {
		t.Fatalf("expected empty first class price, got %s", repository.params.FirstClassPrice)
	}
}

func TestCreateTrainRejectsInvalidPrice(t *testing.T) {
	req := validCreateTrainRequest()
	req.SecondClassPrice = "abc"

	_, err := NewService(&fakeRepository{}).Create(context.Background(), req)
	if !errors.Is(err, ErrInvalidTrain) {
		t.Fatalf("expected invalid train error, got %v", err)
	}
}

func validCreateTrainRequest() CreateTrainRequest {
	return CreateTrainRequest{
		TrainNo:              "G101",
		DepartureDate:        "2026-09-20",
		DepartureStationID:   1,
		ArrivalStationID:     2,
		DepartureTime:        "2026-09-20T08:30:00+08:00",
		ArrivalTime:          "2026-09-20T10:10:00+08:00",
		SaleStartTime:        "2026-09-12T10:00:00+08:00",
		FirstClassPrice:      "199.00",
		SecondClassPrice:     "99.00",
		FirstClassSeatCount:  200,
		SecondClassSeatCount: 1000,
	}
}
