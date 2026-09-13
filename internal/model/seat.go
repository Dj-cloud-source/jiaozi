package model

import "time"

type Seat struct {
	ID            uint64     `db:"id"`
	TrainID       uint64     `db:"train_id"`
	SeatClass     string     `db:"seat_class"`
	SeatNo        uint64     `db:"seat_no"`
	Status        string     `db:"status"`
	LockedOrderID *string    `db:"locked_order_id"`
	LockedAt      *time.Time `db:"locked_at"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}
