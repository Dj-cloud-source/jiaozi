package seat

type LockSeatsParams struct {
	TrainID   uint64
	SeatClass string
	SeatNos   []uint64
	OrderID   string
}
