package auth

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

type CreateUserParams struct {
	Phone        string
	PasswordHash string
	Nickname     string
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (model.User, error) {
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (phone, password_hash, nickname) VALUES (?, ?, ?)`,
		params.Phone,
		params.PasswordHash,
		params.Nickname,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.User{}, ErrPhoneAlreadyRegistered
		}
		return model.User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.User{}, err
	}

	return model.User{
		ID:           uint64(id),
		Phone:        params.Phone,
		PasswordHash: params.PasswordHash,
		Nickname:     params.Nickname,
	}, nil
}
