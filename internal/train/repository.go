package train

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
