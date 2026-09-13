package seat

import "errors"

var (
	ErrInvalidSeatCount  = errors.New("invalid seat count")
	ErrInsufficientSeats = errors.New("insufficient seats")
)
