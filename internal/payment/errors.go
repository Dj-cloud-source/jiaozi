package payment

import "errors"

var (
	ErrInvalidPayment       = errors.New("invalid payment")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentStatusInvalid = errors.New("payment status invalid")
)
