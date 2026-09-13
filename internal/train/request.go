package train

type CreateTrainRequest struct {
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
}

type ListTrainsRequest struct {
	DepartureStationID uint64
	ArrivalStationID   uint64
	Date               string
	TrainNo            string
	Page               int
	PageSize           int
	Sort               string
}
