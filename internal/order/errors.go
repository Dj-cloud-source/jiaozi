package order

import "errors"

var (
	ErrInvalidTicketOrder = errors.New("invalid ticket order")
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderStatusInvalid = errors.New("order status invalid")
)
