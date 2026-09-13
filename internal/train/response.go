package train

import "jiaozi/internal/model"

type AdminTrainResponse struct {
	ID                   uint64 `json:"id"`
	TrainNo              string `json:"train_no"`
	DepartureDate        string `json:"departure_date"`
	DepartureStationID   uint64 `json:"departure_station_id"`
	ArrivalStationID     uint64 `json:"arrival_station_id"`
	DepartureTime        string `json:"departure_time"`
	ArrivalTime          string `json:"arrival_time"`
	SaleStartTime        string `json:"sale_start_time"`
	FirstClassPrice      string `json:"first_class_price"`
	SecondClassPrice     string `json:"second_class_price"`
	FirstClassSeatCount  uint64 `json:"first_class_seat_count"`
	SecondClassSeatCount uint64 `json:"second_class_seat_count"`
	Status               string `json:"status"`
}

func NewAdminTrainResponse(train model.Train) AdminTrainResponse {
	return AdminTrainResponse{
		ID:                   train.ID,
		TrainNo:              train.TrainNo,
		DepartureDate:        train.DepartureDate.Format("2006-01-02"),
		DepartureStationID:   train.DepartureStationID,
		ArrivalStationID:     train.ArrivalStationID,
		DepartureTime:        train.DepartureTime.Format("2006-01-02T15:04:05-07:00"),
		ArrivalTime:          train.ArrivalTime.Format("2006-01-02T15:04:05-07:00"),
		SaleStartTime:        train.SaleStartTime.Format("2006-01-02T15:04:05-07:00"),
		FirstClassPrice:      train.FirstClassPrice,
		SecondClassPrice:     train.SecondClassPrice,
		FirstClassSeatCount:  train.FirstClassSeatCount,
		SecondClassSeatCount: train.SecondClassSeatCount,
		Status:               train.Status,
	}
}

type TrainResponse struct {
	ID               uint64          `json:"id"`
	TrainNo          string          `json:"train_no"`
	DepartureDate    string          `json:"departure_date"`
	DepartureStation StationSnapshot `json:"departure_station"`
	ArrivalStation   StationSnapshot `json:"arrival_station"`
	DepartureTime    string          `json:"departure_time"`
	ArrivalTime      string          `json:"arrival_time"`
	SaleStartTime    string          `json:"sale_start_time"`
	Status           string          `json:"status"`
	FirstClass       SeatClassView   `json:"first_class"`
	SecondClass      SeatClassView   `json:"second_class"`
}

type StationSnapshot struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type SeatClassView struct {
	Price          string  `json:"price"`
	AvailableCount *uint64 `json:"available_count"`
}

func NewTrainResponse(train TrainView) TrainResponse {
	return TrainResponse{
		ID:            train.ID,
		TrainNo:       train.TrainNo,
		DepartureDate: train.DepartureDate.Format("2006-01-02"),
		DepartureStation: StationSnapshot{
			ID:   train.DepartureStationID,
			Name: train.DepartureStationName,
		},
		ArrivalStation: StationSnapshot{
			ID:   train.ArrivalStationID,
			Name: train.ArrivalStationName,
		},
		DepartureTime: train.DepartureTime.Format("2006-01-02T15:04:05-07:00"),
		ArrivalTime:   train.ArrivalTime.Format("2006-01-02T15:04:05-07:00"),
		SaleStartTime: train.SaleStartTime.Format("2006-01-02T15:04:05-07:00"),
		Status:        train.Status,
		FirstClass: SeatClassView{
			Price:          train.FirstClassPrice,
			AvailableCount: availableCount(train.Status, train.FirstClassAvailableCount),
		},
		SecondClass: SeatClassView{
			Price:          train.SecondClassPrice,
			AvailableCount: availableCount(train.Status, train.SecondClassAvailableCount),
		},
	}
}

func availableCount(status string, count uint64) *uint64 {
	if status != "ON_SALE" {
		return nil
	}
	return &count
}
