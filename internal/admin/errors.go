package admin

import "errors"

var (
	ErrAdminAlreadyExists        = errors.New("admin already exists")
	ErrInvalidCreateAdminRequest = errors.New("invalid create admin request")
	ErrInvalidLoginRequest       = errors.New("invalid admin login request")
	ErrInvalidUsernameOrPassword = errors.New("invalid admin username or password")
)
