package station

import (
	"context"
	"database/sql"
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

func (r *Repository) ListAll(ctx context.Context) ([]model.Station, error) {
	var stations []model.Station
	err := r.db.SelectContext(
		ctx,
		&stations,
		`SELECT id, name, status, created_at, updated_at
		 FROM stations
		 ORDER BY id ASC`,
	)
	if err != nil {
		return nil, err
	}

	return stations, nil
}

func (r *Repository) FindByID(ctx context.Context, stationID uint64) (model.Station, error) {
	var station model.Station
	err := r.db.GetContext(
		ctx,
		&station,
		`SELECT id, name, status, created_at, updated_at
		 FROM stations
		 WHERE id = ?`,
		stationID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Station{}, ErrStationNotFound
		}
		return model.Station{}, err
	}

	return station, nil
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

func (r *Repository) UpdateName(ctx context.Context, stationID uint64, name string) (model.Station, error) {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE stations
		 SET name = ?
		 WHERE id = ?`,
		name,
		stationID,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.Station{}, ErrStationExists
		}
		return model.Station{}, err
	}

	return r.FindByID(ctx, stationID)
}

func (r *Repository) Disable(ctx context.Context, stationID uint64) (model.Station, error) {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE stations
		 SET status = ?
		 WHERE id = ?
		   AND status = ?`,
		"DISABLED",
		stationID,
		"ACTIVE",
	)
	if err != nil {
		return model.Station{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return model.Station{}, err
	}
	if affected != 1 {
		if _, err := r.FindByID(ctx, stationID); err != nil {
			return model.Station{}, err
		}
		return model.Station{}, ErrStationStatusInvalid
	}

	return r.FindByID(ctx, stationID)
}
