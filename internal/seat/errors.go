package seat

import "errors"

var (
	ErrInvalidSeatCount  = errors.New("invalid seat count")
	ErrInvalidSeatLock   = errors.New("invalid seat lock")
	ErrInsufficientSeats = errors.New("insufficient seats")
)
