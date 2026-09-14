package order

import "time"

type OrderView struct {
	OrderID              string     `db:"order_id"`
	UserID               uint64     `db:"user_id"`
	TrainID              uint64     `db:"train_id"`
	TrainNo              string     `db:"train_no"`
	DepartureStationName string     `db:"departure_station_name"`
	ArrivalStationName   string     `db:"arrival_station_name"`
	DepartureTime        time.Time  `db:"departure_time"`
	PassengerName        string     `db:"passenger_name"`
	PassengerIDCard      string     `db:"passenger_id_card"`
	SeatClass            string     `db:"seat_class"`
	SeatNo               uint64     `db:"seat_no"`
	TicketPrice          string     `db:"ticket_price"`
	Status               string     `db:"status"`
	PaymentDeadline      time.Time  `db:"payment_deadline"`
	CreatedAt            time.Time  `db:"created_at"`
	TicketedAt            *time.Time `db:"ticketed_at"`
	CancelledAt          *time.Time `db:"cancelled_at"`
	ReturnedAt           *time.Time `db:"returned_at"`
	CompletedAt          *time.Time `db:"completed_at"`
}

type AdminOrderQuery struct {
	Status   string
	TrainNo  string
	Page     int
	PageSize int
}
