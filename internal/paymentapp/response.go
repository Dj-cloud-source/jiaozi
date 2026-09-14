package paymentapp

import (
	"time"

	"jiaozi/internal/model"
	"jiaozi/internal/payment"
)

type PaymentResponse struct {
	PaymentID       string                 `json:"payment_id"`
	Status          string                 `json:"status"`
	OriginalAmount  string                 `json:"original_amount"`
	PayableAmount   string                 `json:"payable_amount"`
	RefundedAmount  string                 `json:"refunded_amount"`
	PaymentDeadline string                 `json:"payment_deadline"`
	Provider        string                 `json:"provider"`
	Orders          []PaymentOrderResponse `json:"orders"`
}

type PaymentOrderResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	TicketPrice string `json:"ticket_price"`
}

type MockSuccessResponse struct {
	PaymentID string   `json:"payment_id"`
	Status    string   `json:"status"`
	OrderIDs  []string `json:"order_ids"`
}

type MockFailResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

func NewPaymentResponse(result PaymentDetailResult) PaymentResponse {
	response := NewBasicPaymentResponse(result.Payment)
	response.Orders = newPaymentOrderResponses(result.Orders)
	return response
}

func NewBasicPaymentResponse(payment model.Payment) PaymentResponse {
	return PaymentResponse{
		PaymentID:       payment.ID,
		Status:          payment.Status,
		OriginalAmount:  payment.OriginalAmount,
		PayableAmount:   payment.PayableAmount,
		RefundedAmount:  payment.RefundedAmount,
		PaymentDeadline: formatTime(payment.PaymentDeadline),
		Provider:        payment.Provider,
		Orders:          []PaymentOrderResponse{},
	}
}

func newPaymentOrderResponses(orders []payment.PaymentOrderView) []PaymentOrderResponse {
	response := make([]PaymentOrderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, PaymentOrderResponse{
			OrderID:     order.OrderID,
			Status:      order.Status,
			TicketPrice: order.TicketPrice,
		})
	}

	return response
}

func NewMockSuccessResponse(result MockSuccessResult) MockSuccessResponse {
	return MockSuccessResponse{
		PaymentID: result.Payment.ID,
		Status:    result.Payment.Status,
		OrderIDs:  result.OrderIDs,
	}
}

func NewMockFailResponse(result MockFailResult) MockFailResponse {
	return MockFailResponse{
		PaymentID: result.Payment.ID,
		Status:    result.Payment.Status,
	}
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02T15:04:05-07:00")
}
