package order

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
	"jiaozi/internal/platform/database"
)

type Repository struct {
	executor database.Executor
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
	return &Repository{executor: db}
}

func (r *Repository) WithExecutor(executor database.Executor) *Repository {
	return &Repository{executor: executor}
}

func (r *Repository) Create(ctx context.Context, params CreateTicketOrderParams) (model.TicketOrder, error) {
	_, err := r.executor.ExecContext(
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

func (r *Repository) ListByUser(ctx context.Context, userID uint64) ([]OrderView, error) {
	var orders []OrderView
	err := r.executor.SelectContext(
		ctx,
		&orders,
		orderViewSelectSQL()+`
		 WHERE o.user_id = ?
		 ORDER BY o.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) FindByIDAndUser(ctx context.Context, orderID string, userID uint64) (OrderView, error) {
	var orderView OrderView
	err := r.executor.GetContext(
		ctx,
		&orderView,
		orderViewSelectSQL()+`
		 WHERE o.id = ?
		   AND o.user_id = ?`,
		orderID,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OrderView{}, ErrOrderNotFound
		}
		return OrderView{}, err
	}

	return orderView, nil
}

func orderViewSelectSQL() string {
	return `SELECT
		o.id AS order_id,
		o.user_id,
		o.train_id,
		t.train_no,
		ds.name AS departure_station_name,
		asn.name AS arrival_station_name,
		t.departure_time,
		p.name AS passenger_name,
		p.id_card AS passenger_id_card,
		s.seat_class,
		s.seat_no,
		CAST(o.ticket_price AS CHAR) AS ticket_price,
		o.status,
		o.payment_deadline,
		o.created_at,
		o.ticketed_at,
		o.cancelled_at,
		o.returned_at,
		o.completed_at
	FROM ticket_orders o
	JOIN trains t ON t.id = o.train_id
	JOIN stations ds ON ds.id = t.departure_station_id
	JOIN stations asn ON asn.id = t.arrival_station_id
	JOIN passengers p ON p.id = o.passenger_id
	JOIN seats s ON s.id = o.seat_id`
}
