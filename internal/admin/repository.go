package admin

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

type CreateAdminParams struct {
	Username     string
	PasswordHash string
}

func (r *Repository) Create(ctx context.Context, params CreateAdminParams) (model.Admin, error) {
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO admins (username, password_hash) VALUES (?, ?)`,
		params.Username,
		params.PasswordHash,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.Admin{}, ErrAdminAlreadyExists
		}
		return model.Admin{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Admin{}, err
	}

	return model.Admin{
		ID:           uint64(id),
		Username:     params.Username,
		PasswordHash: params.PasswordHash,
	}, nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (model.Admin, error) {
	var admin model.Admin
	err := r.db.GetContext(
		ctx,
		&admin,
		`SELECT id, username, password_hash, created_at, updated_at FROM admins WHERE username = ?`,
		username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Admin{}, ErrInvalidUsernameOrPassword
		}
		return model.Admin{}, err
	}

	return admin, nil
}
