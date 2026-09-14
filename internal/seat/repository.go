package seat

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
	"jiaozi/internal/platform/database"
)

type Repository struct {
	executor database.Executor
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{executor: db}
}

func (r *Repository) WithExecutor(executor database.Executor) *Repository {
	return &Repository{executor: executor}
}

func (r *Repository) ListAvailableSeatNos(ctx context.Context, trainID uint64, seatClass string) ([]uint64, error) {
	var seatNos []uint64
	err := r.executor.SelectContext(
		ctx,
		&seatNos,
		`SELECT seat_no
		 FROM seats
		 WHERE train_id = ?
		   AND seat_class = ?
		   AND status = ?
		 ORDER BY seat_no ASC`,
		trainID,
		seatClass,
		"AVAILABLE",
	)
	if err != nil {
		return nil, err
	}

	return seatNos, nil
}

func (r *Repository) ListAvailableSeats(ctx context.Context, trainID uint64, seatClass string) ([]model.Seat, error) {
	var seats []model.Seat
	err := r.executor.SelectContext(
		ctx,
		&seats,
		`SELECT id,
		        train_id,
		        seat_class,
		        seat_no,
		        status,
		        locked_order_id,
		        locked_at,
		        created_at,
		        updated_at
		 FROM seats
		 WHERE train_id = ?
		   AND seat_class = ?
		   AND status = ?
		 ORDER BY seat_no ASC`,
		trainID,
		seatClass,
		"AVAILABLE",
	)
	if err != nil {
		return nil, err
	}

	return seats, nil
}

func (r *Repository) LockSeats(ctx context.Context, params LockSeatsParams) (int64, error) {
	query, args, err := sqlx.In(
		`UPDATE seats
		 SET status = ?,
		     locked_order_id = ?,
		     locked_at = ?
		 WHERE train_id = ?
		   AND seat_class = ?
		   AND status = ?
		   AND seat_no IN (?)`,
		"LOCKED",
		params.OrderID,
		time.Now(),
		params.TrainID,
		params.SeatClass,
		"AVAILABLE",
		params.SeatNos,
	)
	if err != nil {
		return 0, err
	}

	result, err := r.executor.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) ReleaseSeats(ctx context.Context, params OrderSeatActionParams) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE seats
		 SET status = ?,
		     locked_order_id = NULL,
		     locked_at = NULL
		 WHERE status = ?
		   AND locked_order_id = ?`,
		"AVAILABLE",
		"LOCKED",
		params.OrderID,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) MarkSeatsSold(ctx context.Context, params OrderSeatActionParams) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE seats
		 SET status = ?
		 WHERE status = ?
		   AND locked_order_id = ?`,
		"SOLD",
		"LOCKED",
		params.OrderID,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
