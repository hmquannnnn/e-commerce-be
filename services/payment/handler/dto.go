package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/payment-service/model"
)

type CreatePaymentRequest struct {
	OrderID       string `json:"order_id" binding:"required"`
	Provider      string `json:"provider" binding:"required,oneof=payos"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=qr_code"`
	Amount        int64  `json:"amount" binding:"required,min=1"`
	Currency      string `json:"currency" binding:"required"`
	ReturnURL     string `json:"return_url"`
	CancelURL     string `json:"cancel_url"`
}

type PaymentResponse struct {
	ID                     uuid.UUID            `json:"id"`
	OrderID                string               `json:"order_id"`
	UserID                 uuid.UUID            `json:"user_id"`
	Provider               model.Provider       `json:"provider"`
	PaymentMethod          model.PaymentMethod  `json:"payment_method"`
	Amount                 int64                `json:"amount"`
	Currency               string               `json:"currency"`
	Status                 model.PaymentStatus  `json:"status"`
	ProviderPaymentID      *string              `json:"provider_payment_id,omitempty"`
	ProviderTransactionRef *string              `json:"provider_transaction_ref,omitempty"`
	CheckoutURL            *string              `json:"checkout_url,omitempty"`
	ExpiresAt              *string              `json:"expires_at,omitempty"`
	CreatedAt              string               `json:"created_at"`
	UpdatedAt              string               `json:"updated_at"`
}

func ToPaymentResponse(p *model.Payment) *PaymentResponse {
	resp := &PaymentResponse{
		ID:                     p.ID,
		OrderID:                p.OrderID,
		UserID:                 p.UserID,
		Provider:               p.Provider,
		PaymentMethod:          p.PaymentMethod,
		Amount:                 p.Amount,
		Currency:               p.Currency,
		Status:                 p.Status,
		ProviderPaymentID:      p.ProviderPaymentID,
		ProviderTransactionRef: p.ProviderTransactionRef,
		CheckoutURL:            p.CheckoutURL,
		CreatedAt:              p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              p.UpdatedAt.Format(time.RFC3339),
	}
	if p.ExpiresAt != nil {
		v := p.ExpiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &v
	}
	return resp
}
