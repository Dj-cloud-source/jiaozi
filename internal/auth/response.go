package auth

type RegisterResponse struct {
	UserID   uint64 `json:"user_id"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}
