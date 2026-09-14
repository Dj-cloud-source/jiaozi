package payment

type PaymentOrderView struct {
	OrderID     string `db:"order_id"`
	Status      string `db:"status"`
	TicketPrice string `db:"ticket_price"`
}
