package model

import "time"

type User struct {
	ID           uint64    `db:"id"`
	Phone        string    `db:"phone"`
	PasswordHash string    `db:"password_hash"`
	Nickname     string    `db:"nickname"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
