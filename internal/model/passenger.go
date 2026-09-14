package model

import "time"

type Passenger struct {
	ID        uint64    `db:"id"`
	UserID    uint64    `db:"user_id"`
	Name      string    `db:"name"`
	IDCard    string    `db:"id_card"`
	CreatedAt time.Time `db:"created_at"`
}
