package seat

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListAvailableSeatNos(ctx context.Context, trainID uint64, seatClass string) ([]uint64, error) {
	var seatNos []uint64
	err := r.db.SelectContext(
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

	result, err := r.db.ExecContext(ctx, r.db.Rebind(query), args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
