package auth

type RegisterResponse struct {
	UserID   uint64 `json:"user_id"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}

type LoginResponse struct {
	AccessToken string        `json:"access_token"`
	ExpiresIn   int64         `json:"expires_in"`
	User        LoginUserData `json:"user"`
}

type LoginUserData struct {
	ID       uint64 `json:"id"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}
