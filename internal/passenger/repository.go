package passenger

import (
	"context"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
	"jiaozi/internal/platform/database"
)

type Repository struct {
	executor database.Executor
}

type CreatePassengerParams struct {
	UserID uint64
	Name   string
	IDCard string
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{executor: db}
}

func (r *Repository) WithExecutor(executor database.Executor) *Repository {
	return &Repository{executor: executor}
}

func (r *Repository) Create(ctx context.Context, params CreatePassengerParams) (model.Passenger, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO passengers (user_id, name, id_card) VALUES (?, ?, ?)`,
		params.UserID,
		params.Name,
		params.IDCard,
	)
	if err != nil {
		return model.Passenger{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Passenger{}, err
	}

	return model.Passenger{
		ID:     uint64(id),
		UserID: params.UserID,
		Name:   params.Name,
		IDCard: params.IDCard,
	}, nil
}

func (r *Repository) HasActiveTicket(ctx context.Context, idCard string, trainID uint64) (bool, error) {
	var count int
	err := r.executor.GetContext(
		ctx,
		&count,
		`SELECT COUNT(1)
		 FROM ticket_orders o
		 JOIN passengers p ON p.id = o.passenger_id
		 WHERE p.id_card = ?
		   AND o.train_id = ?
		   AND o.status IN (?, ?)`,
		idCard,
		trainID,
		"WAITING_PAYMENT",
		"TICKETED",
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
