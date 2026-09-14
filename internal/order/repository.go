package order

import (
	"context"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

type CreateTicketOrderParams struct {
	ID              string
	UserID          uint64
	PassengerID     uint64
	TrainID         uint64
	SeatID          uint64
	TicketPrice     string
	Status          string
	PaymentDeadline string
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, params CreateTicketOrderParams) (model.TicketOrder, error) {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO ticket_orders (
			id,
			user_id,
			passenger_id,
			train_id,
			seat_id,
			ticket_price,
			status,
			payment_deadline
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		params.ID,
		params.UserID,
		params.PassengerID,
		params.TrainID,
		params.SeatID,
		params.TicketPrice,
		params.Status,
		params.PaymentDeadline,
	)
	if err != nil {
		return model.TicketOrder{}, err
	}

	return model.TicketOrder{
		ID:              params.ID,
		UserID:          params.UserID,
		PassengerID:     params.PassengerID,
		TrainID:         params.TrainID,
		SeatID:          params.SeatID,
		TicketPrice:     params.TicketPrice,
		Status:          params.Status,
		PaymentDeadline: parsePaymentDeadline(params.PaymentDeadline),
	}, nil
}
