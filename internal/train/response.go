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
