package station

import (
	"context"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListActive(ctx context.Context) ([]model.Station, error) {
	var stations []model.Station
	err := r.db.SelectContext(
		ctx,
		&stations,
		`SELECT id, name, status, created_at, updated_at
		 FROM stations
		 WHERE status = ?
		 ORDER BY id ASC`,
		"ACTIVE",
	)
	if err != nil {
		return nil, err
	}

	return stations, nil
}
