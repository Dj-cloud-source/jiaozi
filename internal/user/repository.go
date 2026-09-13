package user

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

func (r *Repository) FindByID(ctx context.Context, id uint64) (model.User, error) {
	var user model.User
	err := r.db.GetContext(
		ctx,
		&user,
		`SELECT id, phone, password_hash, nickname, created_at, updated_at FROM users WHERE id = ?`,
		id,
	)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
