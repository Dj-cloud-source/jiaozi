package user

import "time"

type CurrentUserResponse struct {
	ID        uint64    `json:"id"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
}
