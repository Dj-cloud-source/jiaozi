package model

import "time"

type Train struct {
	ID                   uint64    `db:"id"`
	TrainNo              string    `db:"train_no"`
	DepartureDate        time.Time `db:"departure_date"`
	DepartureStationID   uint64    `db:"departure_station_id"`
	ArrivalStationID     uint64    `db:"arrival_station_id"`
	DepartureTime        time.Time `db:"departure_time"`
	ArrivalTime          time.Time `db:"arrival_time"`
	SaleStartTime        time.Time `db:"sale_start_time"`
	FirstClassPrice      string    `db:"first_class_price"`
	SecondClassPrice     string    `db:"second_class_price"`
	FirstClassSeatCount  uint64    `db:"first_class_seat_count"`
	SecondClassSeatCount uint64    `db:"second_class_seat_count"`
	Status               string    `db:"status"`
	CreatedAt            time.Time `db:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"`
}
