package station

import (
	"context"
	"errors"
	"testing"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	stations []model.Station
	created  model.Station
	createErr error
}

func (r *fakeRepository) ListActive(ctx context.Context) ([]model.Station, error) {
	return r.stations, nil
}

func (r *fakeRepository) ListAll(ctx context.Context) ([]model.Station, error) {
	return r.stations, nil
}

func (r *fakeRepository) FindByID(ctx context.Context, stationID uint64) (model.Station, error) {
	for _, station := range r.stations {
		if station.ID == stationID {
			return station, nil
		}
	}

	return model.Station{}, ErrStationNotFound
}

func (r *fakeRepository) Create(ctx context.Context, name string) (model.Station, error) {
	if r.createErr != nil {
		return model.Station{}, r.createErr
	}
	r.created = model.Station{ID: 10001, Name: name, Status: "ACTIVE"}
	return r.created, nil
}

func (r *fakeRepository) UpdateName(ctx context.Context, stationID uint64, name string) (model.Station, error) {
	for _, station := range r.stations {
		if station.ID == stationID {
			station.Name = name
			return station, nil
		}
	}

	return model.Station{}, ErrStationNotFound
}

func TestListActiveReturnsStations(t *testing.T) {
	service := NewService(&fakeRepository{
		stations: []model.Station{
			{ID: 1, Name: "南京南", Status: "ACTIVE"},
			{ID: 2, Name: "上海虹桥", Status: "ACTIVE"},
		},
	})

	stations, err := service.ListActive(context.Background())
	if err != nil {
		t.Fatalf("list active stations failed: %v", err)
	}

	if len(stations) != 2 {
		t.Fatalf("expected 2 stations, got %d", len(stations))
	}
	if stations[0].Name != "南京南" {
		t.Fatalf("unexpected first station: %s", stations[0].Name)
	}
}

func TestListAllReturnsStations(t *testing.T) {
	service := NewService(&fakeRepository{
		stations: []model.Station{
			{ID: 1, Name: "南京南", Status: "ACTIVE"},
			{ID: 2, Name: "旧站", Status: "DISABLED"},
		},
	})

	stations, err := service.ListAll(context.Background())
	if err != nil {
		t.Fatalf("list all stations failed: %v", err)
	}

	if len(stations) != 2 {
		t.Fatalf("expected 2 stations, got %d", len(stations))
	}
	if stations[1].Status != "DISABLED" {
		t.Fatalf("unexpected second station status: %s", stations[1].Status)
	}
}

func TestCreateStationTrimsName(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	station, err := service.Create(context.Background(), CreateStationRequest{
		Name: " 南京南 ",
	})
	if err != nil {
		t.Fatalf("create station failed: %v", err)
	}

	if station.Name != "南京南" {
		t.Fatalf("unexpected station name: %s", station.Name)
	}
	if station.Status != "ACTIVE" {
		t.Fatalf("unexpected station status: %s", station.Status)
	}
}

func TestCreateStationRejectsEmptyName(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.Create(context.Background(), CreateStationRequest{
		Name: " ",
	})
	if !errors.Is(err, ErrInvalidStationName) {
		t.Fatalf("expected invalid station name error, got %v", err)
	}
}

func TestUpdateStationNameTrimsName(t *testing.T) {
	service := NewService(&fakeRepository{
		stations: []model.Station{
			{ID: 1, Name: "南京南", Status: "ACTIVE"},
		},
	})

	station, err := service.UpdateName(context.Background(), 1, UpdateStationRequest{
		Name: " 南京南站 ",
	})
	if err != nil {
		t.Fatalf("update station failed: %v", err)
	}

	if station.Name != "南京南站" {
		t.Fatalf("unexpected station name: %s", station.Name)
	}
}

func TestUpdateStationNameRejectsEmptyName(t *testing.T) {
	_, err := NewService(&fakeRepository{}).UpdateName(context.Background(), 1, UpdateStationRequest{
		Name: " ",
	})
	if !errors.Is(err, ErrInvalidStationName) {
		t.Fatalf("expected invalid station name error, got %v", err)
	}
}
