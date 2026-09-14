package orderapp

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
