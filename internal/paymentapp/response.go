package paymentapp

import (
	"time"

	"jiaozi/internal/model"
)

type PaymentResponse struct {
	PaymentID       string `json:"payment_id"`
	Status          string `json:"status"`
	OriginalAmount  string `json:"original_amount"`
	PayableAmount   string `json:"payable_amount"`
	RefundedAmount  string `json:"refunded_amount"`
	PaymentDeadline string `json:"payment_deadline"`
	Provider        string `json:"provider"`
}

type MockSuccessResponse struct {
	PaymentID string   `json:"payment_id"`
	Status    string   `json:"status"`
	OrderIDs  []string `json:"order_ids"`
}

func NewPaymentResponse(payment model.Payment) PaymentResponse {
	return PaymentResponse{
		PaymentID:       payment.ID,
		Status:          payment.Status,
		OriginalAmount:  payment.OriginalAmount,
		PayableAmount:   payment.PayableAmount,
		RefundedAmount:  payment.RefundedAmount,
		PaymentDeadline: formatTime(payment.PaymentDeadline),
		Provider:        payment.Provider,
	}
}

func NewMockSuccessResponse(result MockSuccessResult) MockSuccessResponse {
	return MockSuccessResponse{
		PaymentID: result.Payment.ID,
		Status:    result.Payment.Status,
		OrderIDs:  result.OrderIDs,
	}
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02T15:04:05-07:00")
}
