package train

import (
	"context"
	"errors"
	"testing"
	"time"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	params CreateTrainParams
	train  model.Train
	trains []TrainView
	opened int64
	stopped int64
	departed int64
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

func (r *fakeRepository) Update(ctx context.Context, trainID uint64, params CreateTrainParams) (model.Train, error) {
	r.params = params
	if r.err != nil {
		return model.Train{}, r.err
	}

	return model.Train{
		ID:                   trainID,
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

func (r *fakeRepository) Publish(ctx context.Context, trainID uint64) (model.Train, error) {
	if r.err != nil {
		return model.Train{}, r.err
	}
	r.train.ID = trainID
	r.train.Status = "WAITING_SALE"
	return r.train, nil
}

func (r *fakeRepository) List(ctx context.Context, query ListTrainQuery) ([]TrainView, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.trains, nil
}

func (r *fakeRepository) AdminList(ctx context.Context, query AdminListTrainQuery) ([]TrainView, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.trains, nil
}

func (r *fakeRepository) FindViewByID(ctx context.Context, trainID uint64) (TrainView, error) {
	if r.err != nil {
		return TrainView{}, r.err
	}
	if len(r.trains) == 0 {
		return TrainView{}, ErrTrainNotFound
	}
	r.trains[0].ID = trainID
	return r.trains[0], nil
}

func (r *fakeRepository) FindByID(ctx context.Context, trainID uint64) (model.Train, error) {
	if r.err != nil {
		return model.Train{}, r.err
	}
	r.train.ID = trainID
	return r.train, nil
}

func (r *fakeRepository) AdminFindViewByID(ctx context.Context, trainID uint64) (TrainView, error) {
	if r.err != nil {
		return TrainView{}, r.err
	}
	if len(r.trains) == 0 {
		return TrainView{}, ErrTrainNotFound
	}
	r.trains[0].ID = trainID
	return r.trains[0], nil
}

func (r *fakeRepository) OpenDueTrains(ctx context.Context, now time.Time) (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.opened, nil
}

func (r *fakeRepository) StopDueTrains(ctx context.Context, now time.Time) (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.stopped, nil
}

func (r *fakeRepository) DepartDueTrains(ctx context.Context, now time.Time) (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.departed, nil
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

func TestPublishTrainReturnsWaitingSale(t *testing.T) {
	service := NewService(&fakeRepository{
		train: model.Train{
			TrainNo: "G101",
		},
	})

	train, err := service.Publish(context.Background(), 10001)
	if err != nil {
		t.Fatalf("publish train failed: %v", err)
	}

	if train.ID != 10001 {
		t.Fatalf("unexpected train id: %d", train.ID)
	}
	if train.Status != "WAITING_SALE" {
		t.Fatalf("unexpected train status: %s", train.Status)
	}
}

func TestUpdateTrainReturnsDraft(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	req := validCreateTrainRequest()
	req.TrainNo = "G102"
	train, err := service.Update(context.Background(), 10001, req)
	if err != nil {
		t.Fatalf("update train failed: %v", err)
	}

	if train.ID != 10001 {
		t.Fatalf("unexpected train id: %d", train.ID)
	}
	if train.Status != "DRAFT" {
		t.Fatalf("unexpected train status: %s", train.Status)
	}
	if repository.params.TrainNo != "G102" {
		t.Fatalf("unexpected train no: %s", repository.params.TrainNo)
	}
}

func TestUpdateTrainRejectsZeroID(t *testing.T) {
	_, err := NewService(&fakeRepository{}).Update(context.Background(), 0, validCreateTrainRequest())
	if !errors.Is(err, ErrTrainNotFound) {
		t.Fatalf("expected train not found error, got %v", err)
	}
}

func TestPublishTrainRejectsZeroID(t *testing.T) {
	_, err := NewService(&fakeRepository{}).Publish(context.Background(), 0)
	if !errors.Is(err, ErrTrainNotFound) {
		t.Fatalf("expected train not found error, got %v", err)
	}
}

func TestListTrainsAcceptsRouteQuery(t *testing.T) {
	service := NewService(&fakeRepository{
		trains: []TrainView{
			{ID: 10001, TrainNo: "G101", Status: "WAITING_SALE"},
		},
	})

	trains, err := service.List(context.Background(), ListTrainsRequest{
		DepartureStationID: 1,
		ArrivalStationID:   2,
		Date:               "2026-09-20",
		Sort:               "time_asc",
	})
	if err != nil {
		t.Fatalf("list trains failed: %v", err)
	}
	if len(trains) != 1 {
		t.Fatalf("expected 1 train, got %d", len(trains))
	}
}

func TestListTrainsRejectsInvalidSort(t *testing.T) {
	_, err := NewService(&fakeRepository{}).List(context.Background(), ListTrainsRequest{
		DepartureStationID: 1,
		ArrivalStationID:   2,
		Date:               "2026-09-20",
		Sort:               "bad_sort",
	})
	if !errors.Is(err, ErrInvalidTrain) {
		t.Fatalf("expected invalid train error, got %v", err)
	}
}

func TestAdminListTrainsAcceptsEmptyQuery(t *testing.T) {
	service := NewService(&fakeRepository{
		trains: []TrainView{
			{ID: 10001, TrainNo: "G101", Status: "DRAFT"},
			{ID: 10002, TrainNo: "G102", Status: "ARCHIVED"},
		},
	})

	trains, err := service.AdminList(context.Background(), AdminListTrainsRequest{})
	if err != nil {
		t.Fatalf("admin list trains failed: %v", err)
	}
	if len(trains) != 2 {
		t.Fatalf("expected 2 trains, got %d", len(trains))
	}
}

func TestAdminListTrainsRejectsInvalidTrainNo(t *testing.T) {
	_, err := NewService(&fakeRepository{}).AdminList(context.Background(), AdminListTrainsRequest{
		TrainNo: "g101",
	})
	if !errors.Is(err, ErrInvalidTrain) {
		t.Fatalf("expected invalid train error, got %v", err)
	}
}

func TestDetailRejectsZeroID(t *testing.T) {
	_, err := NewService(&fakeRepository{}).Detail(context.Background(), 0)
	if !errors.Is(err, ErrTrainNotFound) {
		t.Fatalf("expected train not found error, got %v", err)
	}
}

func TestOpenDueTrainsReturnsAffectedRows(t *testing.T) {
	service := NewService(&fakeRepository{opened: 2})

	opened, err := service.OpenDueTrains(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("open due trains failed: %v", err)
	}
	if opened != 2 {
		t.Fatalf("expected 2 opened trains, got %d", opened)
	}
}

func TestStopDueTrainsReturnsAffectedRows(t *testing.T) {
	service := NewService(&fakeRepository{stopped: 3})

	stopped, err := service.StopDueTrains(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("stop due trains failed: %v", err)
	}
	if stopped != 3 {
		t.Fatalf("expected 3 stopped trains, got %d", stopped)
	}
}

func TestDepartDueTrainsReturnsAffectedRows(t *testing.T) {
	service := NewService(&fakeRepository{departed: 4})

	departed, err := service.DepartDueTrains(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("depart due trains failed: %v", err)
	}
	if departed != 4 {
		t.Fatalf("expected 4 departed trains, got %d", departed)
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
