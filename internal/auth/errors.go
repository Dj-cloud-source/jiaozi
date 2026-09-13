package auth

import "errors"

var (
	ErrPhoneAlreadyRegistered = errors.New("phone already registered")
	ErrInvalidRegisterRequest = errors.New("invalid register request")
)
