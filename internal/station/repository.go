package station

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
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

func (r *Repository) Create(ctx context.Context, name string) (model.Station, error) {
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO stations (name, status) VALUES (?, ?)`,
		name,
		"ACTIVE",
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.Station{}, ErrStationExists
		}
		return model.Station{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Station{}, err
	}

	return model.Station{
		ID:     uint64(id),
		Name:   name,
		Status: "ACTIVE",
	}, nil
}
