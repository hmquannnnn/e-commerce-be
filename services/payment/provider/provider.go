package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/payment-service/model"
)

type CheckoutRequest struct {
	PaymentID      uuid.UUID
	OrderID        string
	UserID         uuid.UUID
	Amount         int64
	Currency       string
	ReturnURL      string
	CancelURL      string
	PaymentMethod  model.PaymentMethod
	Provider       model.Provider
	PayosOrderCode int64
	ProviderExtras map[string]string
}

type CheckoutResponse struct {
	CheckoutURL            string
	ProviderPaymentID      string
	ProviderTransactionRef string
	ExpiresAt              *time.Time
	RawResponse            []byte
}

type CallbackResult struct {
	EventID                string
	PaymentID              uuid.UUID
	PayosOrderCode         int64
	ProviderPaymentID      string
	ProviderTransactionRef string
	Status                 model.PaymentStatus
	RawPayload             []byte
	SignatureValid         bool
}

type PaymentProvider interface {
	CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)
	VerifyCallback(req *http.Request) (*CallbackResult, error)
}
