package admin

type LoginResponse struct {
	AccessToken string         `json:"access_token"`
	ExpiresIn   int64          `json:"expires_in"`
	Admin       LoginAdminData `json:"admin"`
}

type LoginAdminData struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
}
