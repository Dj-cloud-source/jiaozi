package seat

import "errors"

var (
	ErrInvalidSeatCount  = errors.New("invalid seat count")
	ErrInvalidSeatLock   = errors.New("invalid seat lock")
	ErrInvalidSeatAction = errors.New("invalid seat action")
	ErrInsufficientSeats = errors.New("insufficient seats")
)
