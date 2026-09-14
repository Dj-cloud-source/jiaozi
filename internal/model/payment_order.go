package model

import "time"

type PaymentOrder struct {
	ID        uint64    `db:"id"`
	PaymentID string    `db:"payment_id"`
	OrderID   string    `db:"order_id"`
	Amount    string    `db:"amount"`
	CreatedAt time.Time `db:"created_at"`
}
