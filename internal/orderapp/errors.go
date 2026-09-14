package orderapp

import "errors"

var (
	ErrInvalidCreateOrder        = errors.New("invalid create order")
	ErrTrainNotOnSale            = errors.New("train is not on sale")
	ErrPassengerAlreadyHasTicket = errors.New("passenger already has active ticket")
	ErrNoAvailableSeat           = errors.New("no available seat")
	ErrSeatLockFailed            = errors.New("seat lock failed")
	ErrSeatReleaseFailed         = errors.New("seat release failed")
)
