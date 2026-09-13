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

func (r *Repository) List(ctx context.Context, query ListTrainQuery) ([]TrainView, error) {
	sqlQuery := trainViewSelectSQL()
	args := []interface{}{}

	if query.TrainNo != "" {
		sqlQuery += ` WHERE t.train_no = ? AND t.departure_date = ?`
		args = append(args, query.TrainNo, query.Date)
	} else {
		sqlQuery += ` WHERE t.departure_station_id = ? AND t.arrival_station_id = ? AND t.departure_date = ?`
		args = append(args, query.DepartureStationID, query.ArrivalStationID, query.Date)
	}

	sqlQuery += ` AND t.status IN ('WAITING_SALE', 'ON_SALE', 'STOPPED')`
	sqlQuery += trainViewGroupBySQL()
	sqlQuery += orderBySQL(query.Sort)
	sqlQuery += ` LIMIT ? OFFSET ?`
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)

	var trains []TrainView
	err := r.db.SelectContext(ctx, &trains, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	return trains, nil
}

func (r *Repository) FindViewByID(ctx context.Context, trainID uint64) (TrainView, error) {
	var train TrainView
	err := r.db.GetContext(
		ctx,
		&train,
		trainViewSelectSQL()+` WHERE t.id = ? AND t.status IN ('WAITING_SALE', 'ON_SALE', 'STOPPED')`+trainViewGroupBySQL(),
		trainID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TrainView{}, ErrTrainNotFound
		}
		return TrainView{}, err
	}

	return train, nil
}

func trainViewSelectSQL() string {
	return `SELECT
		t.id,
		t.train_no,
		t.departure_date,
		t.departure_station_id,
		ds.name AS departure_station_name,
		t.arrival_station_id,
		asn.name AS arrival_station_name,
		t.departure_time,
		t.arrival_time,
		t.sale_start_time,
		COALESCE(CAST(t.first_class_price AS CHAR), '') AS first_class_price,
		COALESCE(CAST(t.second_class_price AS CHAR), '') AS second_class_price,
		SUM(CASE WHEN s.seat_class = 'FIRST_CLASS' AND s.status = 'AVAILABLE' THEN 1 ELSE 0 END) AS first_class_available_count,
		SUM(CASE WHEN s.seat_class = 'SECOND_CLASS' AND s.status = 'AVAILABLE' THEN 1 ELSE 0 END) AS second_class_available_count,
		t.status
	FROM trains t
	JOIN stations ds ON ds.id = t.departure_station_id
	JOIN stations asn ON asn.id = t.arrival_station_id
	LEFT JOIN seats s ON s.train_id = t.id`
}

func trainViewGroupBySQL() string {
	return ` GROUP BY
		t.id,
		t.train_no,
		t.departure_date,
		t.departure_station_id,
		ds.name,
		t.arrival_station_id,
		asn.name,
		t.departure_time,
		t.arrival_time,
		t.sale_start_time,
		t.first_class_price,
		t.second_class_price,
		t.status`
}

func orderBySQL(sort string) string {
	switch sort {
	case "time_desc":
		return ` ORDER BY t.departure_time DESC`
	case "price_asc":
		return ` ORDER BY LEAST(COALESCE(t.first_class_price, 99999999), COALESCE(t.second_class_price, 99999999)) ASC`
	case "price_desc":
		return ` ORDER BY LEAST(COALESCE(t.first_class_price, 99999999), COALESCE(t.second_class_price, 99999999)) DESC`
	default:
		return ` ORDER BY t.departure_time ASC`
	}
}

func (r *Repository) OpenDueTrains(ctx context.Context, now time.Time) (int64, error) {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE trains
		 SET status = ?
		 WHERE status = ?
		   AND sale_start_time <= ?`,
		"ON_SALE",
		"WAITING_SALE",
		now,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
