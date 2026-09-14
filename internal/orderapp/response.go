package orderapp

import (
	"time"

	"jiaozi/internal/order"
)

type CreateOrderResponse struct {
	PaymentID       string                 `json:"payment_id"`
	PaymentDeadline string                 `json:"payment_deadline"`
	PayableAmount   string                 `json:"payable_amount"`
	Orders          []CreatedOrderResponse `json:"orders"`
}

type CreatedOrderResponse struct {
	OrderID     string            `json:"order_id"`
	Passenger   PassengerSnapshot `json:"passenger"`
	SeatClass   string            `json:"seat_class"`
	SeatNo      uint64            `json:"seat_no"`
	TicketPrice string            `json:"ticket_price"`
	Status      string            `json:"status"`
}

type PassengerSnapshot struct {
	Name         string `json:"name"`
	IDCardMasked string `json:"id_card_masked"`
}

func NewCreateOrderResponse(result CreateOrderResult) CreateOrderResponse {
	return CreateOrderResponse{
		PaymentID:       result.Payment.ID,
		PaymentDeadline: result.Payment.PaymentDeadline.Format("2006-01-02T15:04:05-07:00"),
		PayableAmount:   result.Payment.PayableAmount,
		Orders: []CreatedOrderResponse{
			{
				OrderID: result.Order.ID,
				Passenger: PassengerSnapshot{
					Name:         result.Passenger.Name,
					IDCardMasked: maskIDCard(result.Passenger.IDCard),
				},
				SeatClass:   result.Seat.SeatClass,
				SeatNo:      result.Seat.SeatNo,
				TicketPrice: result.Order.TicketPrice,
				Status:      result.Order.Status,
			},
		},
	}
}

func maskIDCard(idCard string) string {
	if len(idCard) <= 8 {
		return "****"
	}

	return idCard[:4] + "********" + idCard[len(idCard)-4:]
}

type OrderListItemResponse struct {
	OrderID          string `json:"order_id"`
	TrainNo          string `json:"train_no"`
	DepartureStation string `json:"departure_station"`
	ArrivalStation   string `json:"arrival_station"`
	DepartureTime    string `json:"departure_time"`
	PassengerName    string `json:"passenger_name"`
	SeatClass        string `json:"seat_class"`
	SeatNo           uint64 `json:"seat_no"`
	TicketPrice      string `json:"ticket_price"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
}

func NewOrderListResponse(orders []order.OrderView) []OrderListItemResponse {
	response := make([]OrderListItemResponse, 0, len(orders))
	for _, orderView := range orders {
		response = append(response, OrderListItemResponse{
			OrderID:          orderView.OrderID,
			TrainNo:          orderView.TrainNo,
			DepartureStation: orderView.DepartureStationName,
			ArrivalStation:   orderView.ArrivalStationName,
			DepartureTime:    formatTime(orderView.DepartureTime),
			PassengerName:    orderView.PassengerName,
			SeatClass:        orderView.SeatClass,
			SeatNo:           orderView.SeatNo,
			TicketPrice:      orderView.TicketPrice,
			Status:           orderView.Status,
			CreatedAt:        formatTime(orderView.CreatedAt),
		})
	}

	return response
}

type OrderDetailResponse struct {
	OrderID          string            `json:"order_id"`
	TrainID          uint64            `json:"train_id"`
	TrainNo          string            `json:"train_no"`
	DepartureStation string            `json:"departure_station"`
	ArrivalStation   string            `json:"arrival_station"`
	DepartureTime    string            `json:"departure_time"`
	Passenger         PassengerSnapshot `json:"passenger"`
	SeatClass        string            `json:"seat_class"`
	SeatNo           uint64            `json:"seat_no"`
	TicketPrice      string            `json:"ticket_price"`
	Status           string            `json:"status"`
	PaymentDeadline  string            `json:"payment_deadline"`
	CreatedAt        string            `json:"created_at"`
	TicketedAt        *string           `json:"ticketed_at"`
	CancelledAt      *string           `json:"cancelled_at"`
	ReturnedAt       *string           `json:"returned_at"`
	CompletedAt      *string           `json:"completed_at"`
}

type CancelOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func NewCancelOrderResponse(result CancelOrderResult) CancelOrderResponse {
	return CancelOrderResponse{
		OrderID: result.OrderID,
		Status:  result.Status,
	}
}

func NewOrderDetailResponse(orderView order.OrderView) OrderDetailResponse {
	return OrderDetailResponse{
		OrderID:          orderView.OrderID,
		TrainID:          orderView.TrainID,
		TrainNo:          orderView.TrainNo,
		DepartureStation: orderView.DepartureStationName,
		ArrivalStation:   orderView.ArrivalStationName,
		DepartureTime:    formatTime(orderView.DepartureTime),
		Passenger: PassengerSnapshot{
			Name:         orderView.PassengerName,
			IDCardMasked: maskIDCard(orderView.PassengerIDCard),
		},
		SeatClass:       orderView.SeatClass,
		SeatNo:          orderView.SeatNo,
		TicketPrice:     orderView.TicketPrice,
		Status:          orderView.Status,
		PaymentDeadline: formatTime(orderView.PaymentDeadline),
		CreatedAt:       formatTime(orderView.CreatedAt),
		TicketedAt:       formatOptionalTime(orderView.TicketedAt),
		CancelledAt:     formatOptionalTime(orderView.CancelledAt),
		ReturnedAt:      formatOptionalTime(orderView.ReturnedAt),
		CompletedAt:     formatOptionalTime(orderView.CompletedAt),
	}
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
