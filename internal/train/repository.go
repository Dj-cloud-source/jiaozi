package train

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func (r *Repository) Create(ctx context.Context, params CreateTrainParams) (model.Train, error) {
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO trains (
			train_no,
			departure_date,
			departure_station_id,
			arrival_station_id,
			departure_time,
			arrival_time,
			sale_start_time,
			first_class_price,
			second_class_price,
			first_class_seat_count,
			second_class_seat_count,
			status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		params.TrainNo,
		params.DepartureDate,
		params.DepartureStationID,
		params.ArrivalStationID,
		params.DepartureTime,
		params.ArrivalTime,
		params.SaleStartTime,
		nullablePrice(params.FirstClassPrice),
		nullablePrice(params.SecondClassPrice),
		params.FirstClassSeatCount,
		params.SecondClassSeatCount,
		"DRAFT",
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.Train{}, ErrTrainAlreadyExists
		}
		return model.Train{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Train{}, err
	}

	return model.Train{
		ID:                   uint64(id),
		TrainNo:              params.TrainNo,
		DepartureDate:        params.DepartureDate,
		DepartureStationID:   params.DepartureStationID,
		ArrivalStationID:     params.ArrivalStationID,
		DepartureTime:        params.DepartureTime,
		ArrivalTime:          params.ArrivalTime,
		SaleStartTime:        params.SaleStartTime,
		FirstClassPrice:      params.FirstClassPrice,
		SecondClassPrice:     params.SecondClassPrice,
		FirstClassSeatCount:  params.FirstClassSeatCount,
		SecondClassSeatCount: params.SecondClassSeatCount,
		Status:               "DRAFT",
	}, nil
}

func nullablePrice(price string) interface{} {
	if price == "" {
		return nil
	}
	return price
}

func (r *Repository) Publish(ctx context.Context, trainID uint64) (model.Train, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return model.Train{}, err
	}
	defer tx.Rollback()

	train, err := findTrainForUpdate(ctx, tx, trainID)
	if err != nil {
		return model.Train{}, err
	}
	if train.Status != "DRAFT" {
		return model.Train{}, ErrTrainStatusInvalid
	}

	if err := createSeats(ctx, tx, train.ID, "FIRST_CLASS", train.FirstClassSeatCount); err != nil {
		return model.Train{}, err
	}
	if err := createSeats(ctx, tx, train.ID, "SECOND_CLASS", train.SecondClassSeatCount); err != nil {
		return model.Train{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE trains SET status = ? WHERE id = ? AND status = ?`,
		"WAITING_SALE",
		train.ID,
		"DRAFT",
	)
	if err != nil {
		return model.Train{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Train{}, err
	}

	train.Status = "WAITING_SALE"
	return train, nil
}

type trainRow struct {
	ID                   uint64         `db:"id"`
	TrainNo              string         `db:"train_no"`
	DepartureDate        time.Time      `db:"departure_date"`
	DepartureStationID   uint64         `db:"departure_station_id"`
	ArrivalStationID     uint64         `db:"arrival_station_id"`
	DepartureTime        time.Time      `db:"departure_time"`
	ArrivalTime          time.Time      `db:"arrival_time"`
	SaleStartTime        time.Time      `db:"sale_start_time"`
	FirstClassPrice      sql.NullString `db:"first_class_price"`
	SecondClassPrice     sql.NullString `db:"second_class_price"`
	FirstClassSeatCount  uint64         `db:"first_class_seat_count"`
	SecondClassSeatCount uint64         `db:"second_class_seat_count"`
	Status               string         `db:"status"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
}

func findTrainForUpdate(ctx context.Context, tx *sqlx.Tx, trainID uint64) (model.Train, error) {
	var row trainRow
	err := tx.GetContext(
		ctx,
		&row,
		`SELECT
			id,
			train_no,
			departure_date,
			departure_station_id,
			arrival_station_id,
			departure_time,
			arrival_time,
			sale_start_time,
			first_class_price,
			second_class_price,
			first_class_seat_count,
			second_class_seat_count,
			status,
			created_at,
			updated_at
		 FROM trains
		 WHERE id = ?
		 FOR UPDATE`,
		trainID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Train{}, ErrTrainNotFound
		}
		return model.Train{}, err
	}

	return model.Train{
		ID:                   row.ID,
		TrainNo:              row.TrainNo,
		DepartureDate:        row.DepartureDate,
		DepartureStationID:   row.DepartureStationID,
		ArrivalStationID:     row.ArrivalStationID,
		DepartureTime:        row.DepartureTime,
		ArrivalTime:          row.ArrivalTime,
		SaleStartTime:        row.SaleStartTime,
		FirstClassPrice:      row.FirstClassPrice.String,
		SecondClassPrice:     row.SecondClassPrice.String,
		FirstClassSeatCount:  row.FirstClassSeatCount,
		SecondClassSeatCount: row.SecondClassSeatCount,
		Status:               row.Status,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}

func createSeats(ctx context.Context, tx *sqlx.Tx, trainID uint64, seatClass string, count uint64) error {
	for seatNo := uint64(1); seatNo <= count; seatNo++ {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO seats (train_id, seat_class, seat_no, status) VALUES (?, ?, ?, ?)`,
			trainID,
			seatClass,
			seatNo,
			"AVAILABLE",
		)
		if err != nil {
			return err
		}
	}

	return nil
}
