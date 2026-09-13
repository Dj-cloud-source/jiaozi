package station

import "errors"

var (
	ErrInvalidStationName = errors.New("invalid station name")
	ErrStationExists      = errors.New("station already exists")
)
