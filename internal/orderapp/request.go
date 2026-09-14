package orderapp

type CreateOrderHTTPRequest struct {
	TrainID    uint64                 `json:"train_id" binding:"required"`
	SeatClass  string                 `json:"seat_class" binding:"required"`
	Passengers []CreateOrderPassenger `json:"passengers" binding:"required"`
}

type CreateOrderPassenger struct {
	Name   string `json:"name" binding:"required"`
	IDCard string `json:"id_card" binding:"required"`
}
