package admin

import "errors"

var (
	ErrInvalidLoginRequest      = errors.New("invalid admin login request")
	ErrInvalidUsernameOrPassword = errors.New("invalid admin username or password")
)
