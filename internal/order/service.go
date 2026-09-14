package order

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"time"

	"jiaozi/internal/model"
)

var pricePattern = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,2})?$`)

type repository interface {
	Create(ctx context.Context, params CreateTicketOrderParams) (model.TicketOrder, error)
}

type Service struct {
	repository repository
}

type CreateTicketOrderRequest struct {
	UserID          uint64
	PassengerID     uint64
	TrainID         uint64
	SeatID          uint64
	TicketPrice     string
	PaymentDeadline time.Time
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, req CreateTicketOrderRequest) (model.TicketOrder, error) {
	if req.UserID == 0 ||
		req.PassengerID == 0 ||
		req.TrainID == 0 ||
		req.SeatID == 0 ||
		req.TicketPrice == "" ||
		req.PaymentDeadline.IsZero() ||
		!pricePattern.MatchString(req.TicketPrice) {
		return model.TicketOrder{}, ErrInvalidTicketOrder
	}

	id, err := newUUID()
	if err != nil {
		return model.TicketOrder{}, err
	}

	return s.repository.Create(ctx, CreateTicketOrderParams{
		ID:              id,
		UserID:          req.UserID,
		PassengerID:     req.PassengerID,
		TrainID:         req.TrainID,
		SeatID:          req.SeatID,
		TicketPrice:     req.TicketPrice,
		Status:          "WAITING_PAYMENT",
		PaymentDeadline: req.PaymentDeadline.Format("2006-01-02 15:04:05"),
	})
}

func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3],
		b[4], b[5],
		b[6], b[7],
		b[8], b[9],
		b[10], b[11], b[12], b[13], b[14], b[15],
	), nil
}

func parsePaymentDeadline(value string) time.Time {
	deadline, _ := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	return deadline
}
