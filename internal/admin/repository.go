package admin

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
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
