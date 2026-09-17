package station

import "errors"

var (
	ErrInvalidStationName = errors.New("invalid station name")
	ErrStationNotFound    = errors.New("station not found")
	ErrStationExists      = errors.New("station already exists")
)
