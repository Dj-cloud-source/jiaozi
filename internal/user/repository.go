package user

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

func (r *Repository) AdminList(ctx context.Context, page int, pageSize int) ([]model.User, error) {
	var users []model.User
	err := r.db.SelectContext(
		ctx,
		&users,
		`SELECT id, phone, password_hash, nickname, created_at, updated_at
		 FROM users
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		pageSize,
		(page-1)*pageSize,
	)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) AdminFindByID(ctx context.Context, id uint64) (model.User, error) {
	user, err := r.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}

	return user, nil
}

func (r *Repository) UpdatePasswordHash(ctx context.Context, userID uint64, passwordHash string) (int64, error) {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE users
		 SET password_hash = ?
		 WHERE id = ?`,
		passwordHash,
		userID,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
