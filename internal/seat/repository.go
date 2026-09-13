package seat

import (
	"context"

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
