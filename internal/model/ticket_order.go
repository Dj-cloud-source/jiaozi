package model

import "time"

type TicketOrder struct {
	ID              string     `db:"id"`
	UserID          uint64     `db:"user_id"`
	PassengerID     uint64     `db:"passenger_id"`
	TrainID         uint64     `db:"train_id"`
	SeatID          uint64     `db:"seat_id"`
	TicketPrice     string     `db:"ticket_price"`
	Status          string     `db:"status"`
	PaymentDeadline time.Time  `db:"payment_deadline"`
	CreatedAt       time.Time  `db:"created_at"`
	TicketedAt       *time.Time `db:"ticketed_at"`
	CancelledAt      *time.Time `db:"cancelled_at"`
	ReturnedAt       *time.Time `db:"returned_at"`
	CompletedAt      *time.Time `db:"completed_at"`
}
