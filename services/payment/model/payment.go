package model

import (
	"time"

	"github.com/google/uuid"
)

type Provider string

const (
	ProviderPayOS Provider = "payos"
)

type PaymentMethod string

const (
	MethodQRCode PaymentMethod = "qr_code"
)

type PaymentStatus string

const (
	StatusPending        PaymentStatus = "pending"
	StatusRequiresAction PaymentStatus = "requires_action"
	StatusProcessing     PaymentStatus = "processing"
	StatusSucceeded      PaymentStatus = "succeeded"
	StatusFailed         PaymentStatus = "failed"
	StatusCanceled       PaymentStatus = "canceled"
	StatusExpired        PaymentStatus = "expired"
	StatusRefundedPart   PaymentStatus = "refunded_partial"
	StatusRefundedFull   PaymentStatus = "refunded_full"
)

type Payment struct {
	ID                     uuid.UUID
	OrderID                string
	UserID                 uuid.UUID
	Provider               Provider
	PaymentMethod          PaymentMethod
	Amount                 int64
	Currency               string
	Status                 PaymentStatus
	PayosOrderCode         *int64
	ProviderPaymentID      *string
	ProviderTransactionRef *string
	ReturnURL              *string
	CancelURL              *string
	CheckoutURL            *string
	ExpiresAt              *time.Time
	Metadata               []byte
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type PaymentAttempt struct {
	ID                uuid.UUID
	PaymentID         uuid.UUID
	AttemptNo         int
	ProviderSessionID *string
	CheckoutURL       *string
	Status            string
	RawRequest        []byte
	RawResponse       []byte
	ErrorCode         *string
	ErrorMessage      *string
	CreatedAt         time.Time
}

type WebhookEvent struct {
	ID             uuid.UUID
	Provider       Provider
	EventID        string
	SignatureValid bool
	Payload        []byte
	ProcessedAt    *time.Time
	ProcessStatus  string
	Error          *string
	CreatedAt      time.Time
}
