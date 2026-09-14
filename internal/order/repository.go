package order

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func (r *Repository) AdminList(ctx context.Context, query AdminOrderQuery) ([]OrderView, error) {
	sqlQuery := orderViewSelectSQL()
	args := []interface{}{}

	if query.Status != "" || query.TrainNo != "" {
		sqlQuery += ` WHERE 1 = 1`
		if query.Status != "" {
			sqlQuery += ` AND o.status = ?`
			args = append(args, query.Status)
		}
		if query.TrainNo != "" {
			sqlQuery += ` AND t.train_no = ?`
			args = append(args, query.TrainNo)
		}
	}

	sqlQuery += ` ORDER BY o.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)

	var orders []OrderView
	err := r.executor.SelectContext(ctx, &orders, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) AdminFindByID(ctx context.Context, orderID string) (OrderView, error) {
	var orderView OrderView
	err := r.executor.GetContext(
		ctx,
		&orderView,
		orderViewSelectSQL()+`
		 WHERE o.id = ?`,
		orderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OrderView{}, ErrOrderNotFound
		}
		return OrderView{}, err
	}

	return orderView, nil
}

func (r *Repository) CancelWaitingPayment(ctx context.Context, orderID string, userID uint64, cancelledAt time.Time) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE ticket_orders
		 SET status = ?,
		     cancelled_at = ?
		 WHERE id = ?
		   AND user_id = ?
		   AND status = ?`,
		"CANCELLED",
		cancelledAt,
		orderID,
		userID,
		"WAITING_PAYMENT",
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) ListWaitingPaymentIDsByPayment(ctx context.Context, paymentID string, userID uint64) ([]string, error) {
	var orderIDs []string
	err := r.executor.SelectContext(
		ctx,
		&orderIDs,
		`SELECT o.id
		 FROM ticket_orders o
		 JOIN payment_orders po ON po.order_id = o.id
		 WHERE po.payment_id = ?
		   AND o.user_id = ?
		   AND o.status = ?
		 ORDER BY o.created_at ASC`,
		paymentID,
		userID,
		"WAITING_PAYMENT",
	)
	if err != nil {
		return nil, err
	}

	return orderIDs, nil
}

func (r *Repository) MarkPaymentOrdersTicketed(ctx context.Context, paymentID string, userID uint64, ticketedAt time.Time) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE ticket_orders o
		 JOIN payment_orders po ON po.order_id = o.id
		 SET o.status = ?,
		     o.ticketed_at = ?
		 WHERE po.payment_id = ?
		   AND o.user_id = ?
		   AND o.status = ?`,
		"TICKETED",
		ticketedAt,
		paymentID,
		userID,
		"WAITING_PAYMENT",
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) ListExpiredWaitingPaymentIDs(ctx context.Context, query ExpiredOrderQuery) ([]string, error) {
	var orderIDs []string
	err := r.executor.SelectContext(
		ctx,
		&orderIDs,
		`SELECT id
		 FROM ticket_orders
		 WHERE status = ?
		   AND payment_deadline <= ?
		 ORDER BY payment_deadline ASC
		 LIMIT ?`,
		"WAITING_PAYMENT",
		query.Now,
		query.Limit,
	)
	if err != nil {
		return nil, err
	}

	return orderIDs, nil
}

func (r *Repository) CancelExpiredWaitingPayment(ctx context.Context, orderID string, cancelledAt time.Time) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE ticket_orders
		 SET status = ?,
		     cancelled_at = ?
		 WHERE id = ?
		   AND status = ?
		   AND payment_deadline <= ?`,
		"CANCELLED",
		cancelledAt,
		orderID,
		"WAITING_PAYMENT",
		cancelledAt,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) ReturnTicketed(ctx context.Context, orderID string, userID uint64, returnedAt time.Time) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE ticket_orders
		 SET status = ?,
		     returned_at = ?
		 WHERE id = ?
		   AND user_id = ?
		   AND status = ?`,
		"RETURNED",
		returnedAt,
		orderID,
		userID,
		"TICKETED",
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) ListArrivedTicketedIDs(ctx context.Context, query CompletedOrderQuery) ([]string, error) {
	var orderIDs []string
	err := r.executor.SelectContext(
		ctx,
		&orderIDs,
		`SELECT o.id
		 FROM ticket_orders o
		 JOIN trains t ON t.id = o.train_id
		 WHERE o.status = ?
		   AND t.arrival_time <= ?
		 ORDER BY t.arrival_time ASC
		 LIMIT ?`,
		"TICKETED",
		query.Now,
		query.Limit,
	)
	if err != nil {
		return nil, err
	}

	return orderIDs, nil
}

func (r *Repository) CompleteTicketed(ctx context.Context, orderID string, completedAt time.Time) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE ticket_orders
		 SET status = ?,
		     completed_at = ?
		 WHERE id = ?
		   AND status = ?`,
		"COMPLETED",
		completedAt,
		orderID,
		"TICKETED",
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
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
		o.completed_at,
		pay.id AS payment_id,
		pay.status AS payment_status,
		CAST(pay.payable_amount AS CHAR) AS payment_payable_amount,
		pay.provider AS payment_provider
	FROM ticket_orders o
	JOIN trains t ON t.id = o.train_id
	JOIN stations ds ON ds.id = t.departure_station_id
	JOIN stations asn ON asn.id = t.arrival_station_id
	JOIN passengers p ON p.id = o.passenger_id
	JOIN seats s ON s.id = o.seat_id
	LEFT JOIN payment_orders po ON po.order_id = o.id
	LEFT JOIN payments pay ON pay.id = po.payment_id`
}
