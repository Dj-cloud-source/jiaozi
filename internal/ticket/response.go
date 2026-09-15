package ticket

import (
	"time"

	"jiaozi/internal/order"
)

type Response struct {
	OrderID          string    `json:"order_id"`
	Status           string    `json:"status"`
	TrainNo          string    `json:"train_no"`
	DepartureStation string    `json:"departure_station"`
	ArrivalStation   string    `json:"arrival_station"`
	DepartureTime    string    `json:"departure_time"`
	ArrivalTime      string    `json:"arrival_time"`
	Passenger         Passenger `json:"passenger"`
	Seat              Seat      `json:"seat"`
	TicketPrice      string    `json:"ticket_price"`
	TicketedAt       *string   `json:"ticketed_at"`
	Valid            bool      `json:"valid"`
}

type Passenger struct {
	Name         string `json:"name"`
	IDCardMasked string `json:"id_card_masked"`
}

type Seat struct {
	Class  string `json:"class"`
	Number uint64 `json:"number"`
}

func NewResponse(orderView order.OrderView) Response {
	return Response{
		OrderID:          orderView.OrderID,
		Status:           orderView.Status,
		TrainNo:          orderView.TrainNo,
		DepartureStation: orderView.DepartureStationName,
		ArrivalStation:   orderView.ArrivalStationName,
		DepartureTime:    formatTime(orderView.DepartureTime),
		ArrivalTime:      formatTime(orderView.ArrivalTime),
		Passenger: Passenger{
			Name:         orderView.PassengerName,
			IDCardMasked: maskIDCard(orderView.PassengerIDCard),
		},
		Seat: Seat{
			Class:  orderView.SeatClass,
			Number: orderView.SeatNo,
		},
		TicketPrice: orderView.TicketPrice,
		TicketedAt:  formatOptionalTime(orderView.TicketedAt),
		Valid:       orderView.Status == "TICKETED",
	}
}

func maskIDCard(idCard string) string {
	if len(idCard) <= 8 {
		return "****"
	}

	return idCard[:4] + "********" + idCard[len(idCard)-4:]
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02T15:04:05-07:00")
}

func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := formatTime(*value)
	return &formatted
}
