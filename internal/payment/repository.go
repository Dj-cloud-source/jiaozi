package payment

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
	"jiaozi/internal/platform/database"
)

type Repository struct {
	executor database.Executor
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
	return &Repository{executor: db}
}

func (r *Repository) WithExecutor(executor database.Executor) *Repository {
	return &Repository{executor: executor}
}

func (r *Repository) Create(ctx context.Context, params CreatePaymentParams) (model.Payment, error) {
	_, err := r.executor.ExecContext(
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
		_, err := r.executor.ExecContext(
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

func (r *Repository) FindByIDAndUser(ctx context.Context, paymentID string, userID uint64) (model.Payment, error) {
	var payment model.Payment
	err := r.executor.GetContext(
		ctx,
		&payment,
		`SELECT id,
		        user_id,
		        CAST(original_amount AS CHAR) AS original_amount,
		        CAST(payable_amount AS CHAR) AS payable_amount,
		        CAST(refunded_amount AS CHAR) AS refunded_amount,
		        status,
		        payment_deadline,
		        provider,
		        provider_trade_no,
		        created_at,
		        paid_at,
		        updated_at
		 FROM payments
		 WHERE id = ?
		   AND user_id = ?`,
		paymentID,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Payment{}, ErrPaymentNotFound
		}
		return model.Payment{}, err
	}

	return payment, nil
}

func (r *Repository) ListOrders(ctx context.Context, paymentID string, userID uint64) ([]PaymentOrderView, error) {
	var orders []PaymentOrderView
	err := r.executor.SelectContext(
		ctx,
		&orders,
		`SELECT o.id AS order_id,
		        o.status,
		        CAST(o.ticket_price AS CHAR) AS ticket_price
		 FROM payment_orders po
		 JOIN ticket_orders o ON o.id = po.order_id
		 WHERE po.payment_id = ?
		   AND o.user_id = ?
		 ORDER BY o.created_at ASC`,
		paymentID,
		userID,
	)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) StartPay(ctx context.Context, paymentID string, userID uint64) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payments
		 SET status = ?
		 WHERE id = ?
		   AND user_id = ?
		   AND status IN (?, ?)`,
		"PAYING",
		paymentID,
		userID,
		"UNPAID",
		"FAILED",
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) MarkSuccess(ctx context.Context, paymentID string, userID uint64, paidAt time.Time) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payments
		 SET status = ?,
		     paid_at = ?
		 WHERE id = ?
		   AND user_id = ?
		   AND status = ?`,
		"SUCCESS",
		paidAt,
		paymentID,
		userID,
		"PAYING",
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) MarkFailed(ctx context.Context, paymentID string, userID uint64) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payments
		 SET status = ?
		 WHERE id = ?
		   AND user_id = ?
		   AND status = ?`,
		"FAILED",
		paymentID,
		userID,
		"PAYING",
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) AddRefundedAmountForOrder(ctx context.Context, orderID string, userID uint64, amount string) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payments p
		 JOIN payment_orders po ON po.payment_id = p.id
		 JOIN ticket_orders o ON o.id = po.order_id
		 SET p.refunded_amount = p.refunded_amount + ?
		 WHERE o.id = ?
		   AND o.user_id = ?
		   AND p.status = ?`,
		amount,
		orderID,
		userID,
		"SUCCESS",
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) SubtractPayableAmountForOrder(ctx context.Context, orderID string, userID uint64, amount string) (int64, error) {
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE payments p
		 JOIN payment_orders po ON po.payment_id = p.id
		 JOIN ticket_orders o ON o.id = po.order_id
		 SET p.payable_amount = p.payable_amount - ?
		 WHERE o.id = ?
		   AND o.user_id = ?
		   AND p.status IN (?, ?)
		   AND p.payable_amount >= ?`,
		amount,
		orderID,
		userID,
		"UNPAID",
		"FAILED",
		amount,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
