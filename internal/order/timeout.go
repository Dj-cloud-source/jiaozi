package order

import "time"

type ExpiredOrderQuery struct {
	Now   time.Time
	Limit int
}

type CompletedOrderQuery struct {
	Now   time.Time
	Limit int
}
