package user

type AdminListUsersRequest struct {
	Page     int
	PageSize int
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}
