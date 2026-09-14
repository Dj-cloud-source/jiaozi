package model

import "time"

type Payment struct {
	ID              string     `db:"id"`
	UserID          uint64     `db:"user_id"`
	OriginalAmount  string     `db:"original_amount"`
	PayableAmount   string     `db:"payable_amount"`
	RefundedAmount  string     `db:"refunded_amount"`
	Status          string     `db:"status"`
	PaymentDeadline time.Time  `db:"payment_deadline"`
	Provider        string     `db:"provider"`
	ProviderTradeNo *string    `db:"provider_trade_no"`
	CreatedAt       time.Time  `db:"created_at"`
	PaidAt          *time.Time `db:"paid_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}
