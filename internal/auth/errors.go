package auth

import "errors"

var (
	ErrPhoneAlreadyRegistered = errors.New("phone already registered")
	ErrInvalidRegisterRequest = errors.New("invalid register request")
	ErrInvalidLoginRequest    = errors.New("invalid login request")
	ErrInvalidPhoneOrPassword = errors.New("invalid phone or password")
	ErrUnauthorized           = errors.New("unauthorized")
)
