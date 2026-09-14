package payment

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

type CreatePaymentParams struct {
	ID              string
	UserID          uint64
	OriginalAmount  string
	PayableAmount   string
	RefundedAmount  string
	Status          string
	PaymentDeadline string
	Provider        string
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, params CreatePaymentParams) (model.Payment, error) {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO payments (
			id,
			user_id,
			original_amount,
			payable_amount,
			refunded_amount,
			status,
			payment_deadline,
			provider
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		params.ID,
		params.UserID,
		params.OriginalAmount,
		params.PayableAmount,
		params.RefundedAmount,
		params.Status,
		params.PaymentDeadline,
		params.Provider,
	)
	if err != nil {
		return model.Payment{}, err
	}

	return model.Payment{
		ID:              params.ID,
		UserID:          params.UserID,
		OriginalAmount:  params.OriginalAmount,
		PayableAmount:   params.PayableAmount,
		RefundedAmount:  params.RefundedAmount,
		Status:          params.Status,
		PaymentDeadline: parsePaymentDeadline(params.PaymentDeadline),
		Provider:        params.Provider,
	}, nil
}

func parsePaymentDeadline(value string) time.Time {
	deadline, _ := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	return deadline
}

func (r *Repository) CreateOrderLinks(ctx context.Context, links []PaymentOrderLink) error {
	for _, link := range links {
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO payment_orders (payment_id, order_id, amount) VALUES (?, ?, ?)`,
			link.PaymentID,
			link.OrderID,
			link.Amount,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
