package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/payment-service/model"
	"github.com/lib/pq"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrDuplicateEvent  = errors.New("duplicate webhook event")
)

// jsonbParam converts a []byte to a driver-safe value for JSONB columns.
// A nil/empty slice becomes SQL NULL; otherwise the bytes are cast to string
// so lib/pq sends them in text format instead of bytea hex encoding.
func jsonbParam(data []byte) interface{} {
	if len(data) == 0 {
		return nil
	}
	return string(data)
}

type PaymentRepository interface {
	NextPayosOrderCode(ctx context.Context) (int64, error)
	CreatePaymentAndAttempt(ctx context.Context, payment *model.Payment, attempt *model.PaymentAttempt) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Payment, error)
	GetByPayosOrderCode(ctx context.Context, orderCode int64) (*model.Payment, error)
	CancelPending(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ApplyCallback(ctx context.Context, provider model.Provider, result *ApplyCallbackParams) error
}

type ApplyCallbackParams struct {
	EventID                string
	PaymentID              uuid.UUID
	Status                 model.PaymentStatus
	ProviderPaymentID      string
	ProviderTransactionRef string
	RawPayload             []byte
	SignatureValid         bool
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) NextPayosOrderCode(ctx context.Context) (int64, error) {
	var code int64
	err := r.db.QueryRowContext(ctx, `SELECT nextval('payos_order_code_seq')`).Scan(&code)
	if err != nil {
		return 0, fmt.Errorf("nextval payos_order_code_seq: %w", err)
	}
	return code, nil
}

func (r *paymentRepository) CreatePaymentAndAttempt(ctx context.Context, payment *model.Payment, attempt *model.PaymentAttempt) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments (
			id, order_id, user_id, provider, payment_method, amount, currency, status,
			payos_order_code, provider_payment_id, provider_transaction_ref, return_url, cancel_url, checkout_url, expires_at, metadata
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
		)
	`,
		payment.ID, payment.OrderID, payment.UserID, payment.Provider, payment.PaymentMethod, payment.Amount, payment.Currency, payment.Status,
		payment.PayosOrderCode, payment.ProviderPaymentID, payment.ProviderTransactionRef, payment.ReturnURL, payment.CancelURL, payment.CheckoutURL, payment.ExpiresAt,
		jsonbParam(payment.Metadata),
	)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment_attempts (
			id, payment_id, attempt_no, provider_session_id, checkout_url, status, raw_request, raw_response, error_code, error_message
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10
		)
	`,
		attempt.ID, attempt.PaymentID, attempt.AttemptNo, attempt.ProviderSessionID, attempt.CheckoutURL, attempt.Status,
		jsonbParam(attempt.RawRequest), jsonbParam(attempt.RawResponse),
		attempt.ErrorCode, attempt.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("insert payment attempt: %w", err)
	}
	return tx.Commit()
}

const paymentColumns = `id, order_id, user_id, provider, payment_method, amount, currency, status, payos_order_code, provider_payment_id, provider_transaction_ref, return_url, cancel_url, checkout_url, expires_at, metadata, created_at, updated_at`

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	return r.getOne(ctx, `SELECT `+paymentColumns+` FROM payments WHERE id = $1`, id)
}

func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID string) (*model.Payment, error) {
	return r.getOne(ctx, `SELECT `+paymentColumns+` FROM payments WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1`, orderID)
}

func (r *paymentRepository) GetByPayosOrderCode(ctx context.Context, orderCode int64) (*model.Payment, error) {
	return r.getOne(ctx, `SELECT `+paymentColumns+` FROM payments WHERE payos_order_code = $1`, orderCode)
}

func (r *paymentRepository) getOne(ctx context.Context, query string, arg interface{}) (*model.Payment, error) {
	p := &model.Payment{}
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.Provider, &p.PaymentMethod, &p.Amount, &p.Currency, &p.Status,
		&p.PayosOrderCode, &p.ProviderPaymentID, &p.ProviderTransactionRef, &p.ReturnURL, &p.CancelURL, &p.CheckoutURL, &p.ExpiresAt, &p.Metadata, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("query payment: %w", err)
	}
	return p, nil
}

func (r *paymentRepository) CancelPending(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE payments
		SET status='canceled', updated_at=NOW()
		WHERE id=$1 AND user_id=$2 AND status IN ('pending','requires_action','processing')
	`, id, userID)
	if err != nil {
		return fmt.Errorf("cancel payment: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

const callbackAttemptNo = 999

func (r *paymentRepository) ApplyCallback(ctx context.Context, provider model.Provider, params *ApplyCallbackParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment_webhook_events (id, provider, event_id, signature_valid, payload, process_status, processed_at)
		VALUES ($1,$2,$3,$4,$5,'processed',NOW())
	`, uuid.New(), provider, params.EventID, params.SignatureValid, jsonbParam(params.RawPayload))
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateEvent
		}
		return fmt.Errorf("insert webhook event: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE payments
		SET status=$1,
			provider_payment_id = COALESCE(NULLIF($2, ''), provider_payment_id),
			provider_transaction_ref = COALESCE(NULLIF($3, ''), provider_transaction_ref),
			updated_at=NOW()
		WHERE id=$4
	`, params.Status, params.ProviderPaymentID, params.ProviderTransactionRef, params.PaymentID)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment_attempts (
			id, payment_id, attempt_no, provider_session_id, checkout_url, status, raw_request, raw_response, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`,
		uuid.New(), params.PaymentID, callbackAttemptNo, params.ProviderPaymentID, nil, string(params.Status), nil, jsonbParam(params.RawPayload), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("insert callback attempt: %w", err)
	}
	return tx.Commit()
}
