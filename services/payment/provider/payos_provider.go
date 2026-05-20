package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hmquannnnn/e-commerce/payment-service/config"
	"github.com/hmquannnnn/e-commerce/payment-service/model"
	"github.com/payOSHQ/payos-lib-golang/v2"
)

type PayOSProvider struct {
	client *payos.PayOS
	cfg    config.PayOSConfig
}

var ErrOrderCodeExists = errors.New("payos order code already exists")

func NewPayOSProvider(cfg config.PayOSConfig) (*PayOSProvider, error) {
	client, err := payos.NewPayOS(&payos.PayOSOptions{
		ClientId:    cfg.ClientID,
		ApiKey:      cfg.ApiKey,
		ChecksumKey: cfg.ChecksumKey,
	})
	if err != nil {
		return nil, fmt.Errorf("init payos client: %w", err)
	}
	return &PayOSProvider{client: client, cfg: cfg}, nil
}

func (p *PayOSProvider) CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error) {
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = p.cfg.ReturnURL
	}
	cancelURL := req.CancelURL
	if cancelURL == "" {
		cancelURL = p.cfg.CancelURL
	}

	description := fmt.Sprintf("DH %d", req.PayosOrderCode)
	if len(description) > 25 {
		description = description[:25]
	}

	result, err := p.client.PaymentRequests.Create(ctx, payos.CreatePaymentLinkRequest{
		OrderCode:   req.PayosOrderCode,
		Amount:      int(req.Amount),
		Description: description,
		ReturnUrl:   returnURL,
		CancelUrl:   cancelURL,
	})
	if err != nil {
		if isOrderCodeExistsError(err) {
			return nil, fmt.Errorf("%w: %v", ErrOrderCodeExists, err)
		}
		return nil, fmt.Errorf("create payos payment link: %w", err)
	}

	raw, _ := json.Marshal(result)
	return &CheckoutResponse{
		CheckoutURL:       result.CheckoutUrl,
		ProviderPaymentID: result.PaymentLinkId,
		RawResponse:       raw,
	}, nil
}

func isOrderCodeExistsError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "code 231")
}

func (p *PayOSProvider) VerifyCallback(req *http.Request) (*CallbackResult, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read webhook body: %w", err)
	}

	var webhookBody map[string]interface{}
	if err := json.Unmarshal(body, &webhookBody); err != nil {
		return nil, fmt.Errorf("parse webhook json: %w", err)
	}

	verifiedData, err := p.client.Webhooks.VerifyData(req.Context(), webhookBody)
	if err != nil {
		return nil, fmt.Errorf("invalid payos signature: %w", err)
	}

	dataMap, ok := verifiedData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected webhook data type")
	}

	var orderCode int64
	if v, ok := dataMap["orderCode"].(float64); ok {
		orderCode = int64(v)
	}
	paymentLinkID, _ := dataMap["paymentLinkId"].(string)
	reference, _ := dataMap["reference"].(string)

	status := mapPayOSStatus(webhookBody)

	eventID := reference
	if eventID == "" {
		eventID = fmt.Sprintf("%d-%s", orderCode, paymentLinkID)
	}

	return &CallbackResult{
		EventID:                eventID,
		PayosOrderCode:         orderCode,
		ProviderPaymentID:      paymentLinkID,
		ProviderTransactionRef: reference,
		Status:                 status,
		RawPayload:             body,
		SignatureValid:         true,
	}, nil
}

func mapPayOSStatus(body map[string]interface{}) model.PaymentStatus {
	code, _ := body["code"].(string)
	if code == "00" {
		return model.StatusSucceeded
	}
	// PayOS samples sometimes omit top-level code; rely on success when code is absent.
	if s, ok := body["success"].(bool); ok && s && code == "" {
		return model.StatusSucceeded
	}
	return model.StatusFailed
}
