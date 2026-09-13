package station

import (
	"context"
	"testing"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	stations []model.Station
}

func (r *fakeRepository) ListActive(ctx context.Context) ([]model.Station, error) {
	return r.stations, nil
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
