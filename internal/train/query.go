package train

import "time"

type ListTrainQuery struct {
	DepartureStationID uint64
	ArrivalStationID   uint64
	Date               time.Time
	TrainNo            string
	Page               int
	PageSize           int
	Sort               string
}

type AdminListTrainQuery struct {
	TrainNo  string
	Page     int
	PageSize int
}

type TrainView struct {
	ID                         uint64    `db:"id"`
	TrainNo                    string    `db:"train_no"`
	DepartureDate              time.Time `db:"departure_date"`
	DepartureStationID         uint64    `db:"departure_station_id"`
	DepartureStationName       string    `db:"departure_station_name"`
	ArrivalStationID           uint64    `db:"arrival_station_id"`
	ArrivalStationName         string    `db:"arrival_station_name"`
	DepartureTime              time.Time `db:"departure_time"`
	ArrivalTime                time.Time `db:"arrival_time"`
	SaleStartTime              time.Time `db:"sale_start_time"`
	FirstClassPrice            string    `db:"first_class_price"`
	SecondClassPrice           string    `db:"second_class_price"`
	FirstClassAvailableCount   uint64    `db:"first_class_available_count"`
	SecondClassAvailableCount  uint64    `db:"second_class_available_count"`
	Status                     string    `db:"status"`
}
